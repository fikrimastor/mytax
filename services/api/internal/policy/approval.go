package policy

import (
 "errors"
 "strings"
)
// ReleaseRecord is separate from a caller-supplied Package.Status.
// A registered release requires human review and immutable source evidence.
type ReleaseRecord struct {
 PackageVersion string `json:"package_version"`
 AssessmentYear int `json:"assessment_year"`
 ReviewerID string `json:"reviewer_id"`
 ReviewedAt string `json:"reviewed_at"`
 SourceSnapshotSHA256 string `json:"source_snapshot_sha256"`
 ApprovalReference string `json:"approval_reference"`
}
func ValidateRelease(p Package, a ReleaseRecord)error{
 if p.Status!=Approved{return errors.New("policy package is not marked approved")}
 if p.Version==""||p.Version!=a.PackageVersion||p.AssessmentYear!=a.AssessmentYear{return errors.New("release identity mismatch")}
 if a.ReviewerID==""||a.ReviewedAt==""||a.ApprovalReference==""||len(a.SourceSnapshotSHA256)!=64{return errors.New("release audit evidence missing")}
 for _,ch:=range a.SourceSnapshotSHA256{if !strings.ContainsRune("0123456789abcdefABCDEF",ch){return errors.New("invalid SHA256")}}
 if len(p.Rules)==0{return errors.New("empty release")}
 for _,r:=range p.Rules{
  if len(r.Sources)==0{return errors.New("rule source missing")}
  for _,s:=range r.Sources{if !strings.HasPrefix(s.URL,"https://www.hasil.gov.my/")||s.Section==""||s.RetrievedOn==""{return errors.New("source provenance incomplete")}}
 }
 return nil
}
// No active, independently reviewed production release is registered.
func ActivePackage(year int)(Package,error){
 return Package{},errors.New("no active approved LHDN policy package for assessment year")
}
