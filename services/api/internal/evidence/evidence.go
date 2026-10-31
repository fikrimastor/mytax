// Package evidence provides an in-memory, metadata-only duplicate and audit preview.
// No real document bytes, identifiers, or personal taxpayer data should be supplied.
package evidence

import (
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "sort"
 "strings"
)
type Item struct {
 ID string `json:"id"`
 ClaimID string `json:"claim_id"`
 DocumentDigest string `json:"document_digest"`
 MerchantKey string `json:"merchant_key"`
 InvoiceKey string `json:"invoice_key"`
 AmountSen int64 `json:"amount_sen"`
}
type Finding struct { EvidenceID string `json:"evidence_id"`; DuplicateOf string `json:"duplicate_of"`; Reason string `json:"reason"` }
type AuditEvent struct { Sequence int `json:"sequence"`; Action string `json:"action"`; SubjectID string `json:"subject_id"`; PreviousHash string `json:"previous_hash"`; Hash string `json:"hash"` }
type Result struct { Status string `json:"status"`; Findings []Finding `json:"findings"`; Audit []AuditEvent `json:"audit"`; FilingReady bool `json:"filing_ready"` }
func hash(s string)string{h:=sha256.Sum256([]byte(s));return hex.EncodeToString(h[:])}
func validDigest(s string)bool{if len(s)!=64{return false};for _,r:=range s{if !strings.ContainsRune("0123456789abcdef",r){return false}};return true}
func Inspect(items []Item)(Result,error){
 if len(items)>1000{return Result{},errors.New("too many evidence items")}
 ordered:=append([]Item(nil),items...);sort.Slice(ordered,func(i,j int)bool{return ordered[i].ID<ordered[j].ID})
 out:=Result{Status:"METADATA_ONLY_PREVIEW",Findings:[]Finding{},Audit:[]AuditEvent{},FilingReady:false}
 ids:=map[string]bool{};digests:=map[string]string{};invoices:=map[string]string{};previous:=strings.Repeat("0",64)
 for _,v:=range ordered{
  if v.ID==""||v.ClaimID==""||ids[v.ID]||v.AmountSen<=0||!validDigest(v.DocumentDigest){return Result{},errors.New("invalid or duplicate evidence metadata")}
  ids[v.ID]=true
  if prior,ok:=digests[v.DocumentDigest];ok{out.Findings=append(out.Findings,Finding{v.ID,prior,"IDENTICAL_DOCUMENT_DIGEST"})}else{digests[v.DocumentDigest]=v.ID}
  if v.MerchantKey!=""&&v.InvoiceKey!=""{
   key:=v.MerchantKey+"\x00"+v.InvoiceKey
   if prior,ok:=invoices[key];ok {out.Findings=append(out.Findings,Finding{v.ID,prior,"SAME_MERCHANT_INVOICE"})}else{invoices[key]=v.ID}
  }
  signature:=previous+"|EVIDENCE_SCREENED|"+v.ID+"|"+v.ClaimID+"|"+v.DocumentDigest
  event:=AuditEvent{Sequence:len(out.Audit)+1,Action:"EVIDENCE_SCREENED",SubjectID:v.ID,PreviousHash:previous,Hash:hash(signature)}
  out.Audit=append(out.Audit,event);previous=event.Hash
 }
 return out,nil
}
// VerifyAudit detects modifications of the serialized audit chain.
func VerifyAudit(items []Item,events []AuditEvent)bool{
 ordered:=append([]Item(nil),items...);sort.Slice(ordered,func(i,j int)bool{return ordered[i].ID<ordered[j].ID})
 if len(ordered)!=len(events){return false}
 prev:=strings.Repeat("0",64)
 for i,v:=range ordered{
  e:=events[i]
  expected:=hash(prev+"|EVIDENCE_SCREENED|"+v.ID+"|"+v.ClaimID+"|"+v.DocumentDigest)
  if e.Sequence!=i+1||e.Action!="EVIDENCE_SCREENED"||e.SubjectID!=v.ID||e.PreviousHash!=prev||e.Hash!=expected{return false}
  prev=e.Hash
 }
 return true
}
