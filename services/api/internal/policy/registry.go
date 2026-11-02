package policy

// RegistryEntry is a descriptive rule inventory, not an eligibility approval.
// Amounts are in integer sen. Multiple entries may be subject to overlapping
// legal restrictions not fully represented in this catalog.
type RegistryEntry struct {
 ID string `json:"id"`
 Label string `json:"label"`
 MaxSen int64 `json:"max_sen"`
 ParentID string `json:"parent_id,omitempty"`
 Conditional bool `json:"conditional"`
 Note string `json:"note"`
 SourceURL string `json:"source_url"`
 Status string `json:"status"`
}

const OfficialReliefSource = "https://www.hasil.gov.my/individu/pelepasan-cukai/"

// ReliefRegistry2025 is a read-only, draft catalog for discovery/review.
// It MUST NOT be mapped to Approved rules without detailed legal sign-off.
func ReliefRegistry2025() []RegistryEntry {
 data := []struct{id,label,parent,note string;rm int64}{
  {"SELF","Individual and dependent relatives","","Resident individual; subject to assessment context",9000},
  {"PARENTS","Parents/grandparents expenses","","Qualifying expenses and medical confirmation",8000},
  {"PARENTS_CHECKUP","Parents/grandparents checkup and vaccination","PARENTS","Shared sublimit",1000},
  {"DISABILITY_EQUIPMENT","Basic supporting equipment","","Qualifying disability and beneficiary",6000},
  {"DISABLED_SELF","Disabled individual","","Qualifying disability certification",7000},
  {"SELF_EDUCATION","Self education","","Subject to eligible education/course conditions",7000},
  {"SELF_UPSKILL","Up-skilling or self-enhancement course","SELF_EDUCATION","Shared sublimit",2000},
  {"MEDICAL","Self/spouse/child medical","","Multiple sublimits, qualifying diagnoses and expenses",10000},
  {"MEDICAL_VACCINATION","Vaccination","MEDICAL","Subject to qualifying vaccines",1000},
  {"MEDICAL_DENTAL","Dental examination/treatment","MEDICAL","Subject to permitted services",1000},
  {"MEDICAL_SCREENING","Health examination and permitted tests","MEDICAL","Shared sublimit; medical devices may have separate conditions",1000},
  {"MEDICAL_LEARNING_DISABILITY","Child learning disability","MEDICAL","Qualifying child and services",6000},
  {"LIFESTYLE","Lifestyle","","Qualifying books, electronics, Internet and courses",2500},
  {"SPORTS","Additional sports lifestyle","","Eligible equipment, facilities, competitions and gym",1000},
  {"EV_COMPOST","EV charging and household compost equipment","","Qualifying eligible expenses and shared limits",2500},
  {"CHILDCARE","Registered childcare/kindergarten fees","","Eligible child age, provider and spouse claim allocation",3000},
  {"BREASTFEEDING","Breastfeeding equipment","","Female taxpayer, eligible child and once every 2 YA",1000},
  {"SSPN","SSPN net savings","","Deposits less withdrawals; spouse allocation",8000},
  {"LIFE_EPF","Life insurance and EPF","","Distinct component caps must be implemented",7000},
  {"EDU_MED_INSURANCE","Education and medical insurance","","Qualifying premium and beneficiary",4000},
  {"PRS","Private retirement schemes/deferred annuity","","Qualifying contributions",3000},
  {"SOCSO","SOCSO contributions","","Qualifying payments",350},
  {"SPOUSE","Spouse/alimony","","Qualifying joint/spouse status",4000},
  {"DISABLED_SPOUSE","Disabled spouse","","Qualifying disability and spouse relief",6000},
  {"CHILD_UNDER_18","Unmarried child under 18","","Per-child amount, claim allocation",2000},
  {"CHILD_STUDY_BASIC","Unmarried child 18+ studying","","Education conditions; alternative to higher category",2000},
  {"CHILD_STUDY_HIGHER","Unmarried child 18+ higher education","","Education conditions; alternative to basic",8000},
  {"DISABLED_CHILD","Unmarried disabled child","","Per-child amount; additional conditions",8000},
  {"DISABLED_CHILD_STUDY","Additional disabled child higher education","DISABLED_CHILD","Additional qualifying education relief",8000},
  {"FIRST_HOME_INTEREST_LOW","First home interest: home up to RM500k","","SPA dates, price, first-home and other restrictions",7000},
  {"FIRST_HOME_INTEREST_MID","First home interest: RM500k–RM750k","","Alternative tier; not additive with low tier",5000},
 }
 out:=make([]RegistryEntry,0,len(data))
 for _,e:=range data {
  note:=e.note
  if f,ok:=act53Findings[e.id];ok{note+=" | "+f.Kind+": "+f.Text}
  out=append(out,RegistryEntry{ID:e.id,Label:e.label,MaxSen:e.rm*100,ParentID:e.parent,Conditional:true,Note:note,SourceURL:OfficialReliefSource,Status:"DRAFT_REVIEW_REQUIRED"})
 }
 return out
}

// act53Finding records how a registry entry compares with the Income Tax Act 1967 (Act 53)
// updated texts (as at 21 May 2024 and 1 January 2026; neither is an authentic text) and with
// the Finance Act 2024 (Act 862) and Finance Act 2025 (Act 874) texts supplied by the repository
// owner. MaxSen stays the LHDN YA2025 table value. Nothing here is tax-professional reviewed.
// See docs/ACT53_RECONCILIATION_YA2025.md and docs/sources/act53.
type act53Finding struct{Kind,Text string}

const (
 // The Act and the amending Finance Act texts give the same value and place it in YA2025.
 Act53Confirmed="ACT53_CONFIRMED_YA2025"
 // Value equals a later Act text, but the supplied texts do not show whether it applied in YA2025. No entry uses this now.
 Act53YAUncertain="ACT53_YA_UNCERTAIN"
 // The supplied Act texts give a different value for YA2025. Unresolved. No entry uses this now.
 Act53Differs="ACT53_DIFFERS"
 // Amount agrees but scope or conditions need care or are not modelled.
 Act53Scope="ACT53_SCOPE"
)

var act53Findings=map[string]act53Finding{
 "SSPN":{Act53Confirmed,"Finance Act 2024 s.9(e) replaces s.46(1)(k): RM8,000 deposited for his child for YA2025, 2026 and 2027, less withdrawals (a withdrawal for the child's higher education is not deducted); one spouse only if assessed separately."},
 "EV_COMPOST":{Act53Confirmed,"Finance Act 2024 s.9(h), s.46(1)(v): EV charging YA2023-2027; food waste compost machine YA2025-2027, claimable ONCE; total RM2,500. Food waste grinder and CCTV start in YA2026 (Finance Act 2025) and are not YA2025."},
 "DISABLED_SELF":{Act53Confirmed,"Finance Act 2024 s.9(b): s.46(1)(e) raised from RM6,000 to RM7,000 for YA2025 onward."},
 "DISABLED_SPOUSE":{Act53Confirmed,"Finance Act 2024 s.11: s.47(1)(b) raised from RM5,000 to RM6,000 for YA2025 onward. Finance Act 2024 s.8 makes the same change in s.45A(1) for a wife claiming a disabled husband (RM4,000 plus a further RM6,000)."},
 "DISABLED_CHILD":{Act53Confirmed,"Finance Act 2024 s.12: s.48(2)(b) raised from RM6,000 to RM8,000 for YA2025 onward."},
 "EDU_MED_INSURANCE":{Act53Confirmed,"Finance Act 2024 s.13: s.49(1B) raised from RM3,000 to RM4,000 for YA2025 onward."},
 "MEDICAL_LEARNING_DISABILITY":{Act53Confirmed,"Finance Act 2024 s.9(d): s.46(1)(ha) raised from RM4,000 to RM6,000 for YA2025 onward, part of the RM10,000 in s.46(1)(g). Finance Act 2025 s.6(a)(iii) raises it to RM10,000 from YA2026 only."},
 "CHILDCARE":{Act53Confirmed,"Finance Act 2024 s.9(f) extends the further RM1,000 in s.46(1)(r) proviso (c) from YA2024 to YA2027, so YA2025 is RM2,000 plus RM1,000. Child up to 6 at a registered centre or kindergarten. Finance Act 2025 s.6(a)(iv) (YA2026 onward) makes it a flat RM3,000 and adds care centres for a child up to 12."},
 "FIRST_HOME_INTEREST_LOW":{Act53Confirmed,"Finance Act 2024 s.10 replaces s.46B, in force 1 Jan 2025: citizen and resident; first residential property, one unit, occupied as residence; SPA 1 Jan 2025 to 31 Dec 2027; no income from the property; RM7,000 (price up to RM500,000) for 3 consecutive basis years; shared by formula A x B / C between co-claimants."},
 "FIRST_HOME_INTEREST_MID":{Act53Confirmed,"Finance Act 2024 s.10 replaces s.46B, in force 1 Jan 2025: same conditions as the low tier; RM5,000 where price is above RM500,000 and up to RM750,000."},
 "MEDICAL_SCREENING":{Act53Confirmed,"Finance Act 2024 s.9(c), s.46(1)(h) from YA2025: complete medical examination or disease detection test fee, self-testing medical device registered under the Medical Device Act 2012 (not for business), mental health examination; part of the RM10,000. COVID-19 tests are no longer named separately in (h)(ii) although the pasted LHDN table still names them."},
 "PARENTS":{Act53Confirmed,"Finance Act 2024 s.9(a) from YA2025: s.46(1)(c) covers parents or grandparents resident in Malaysia, and a complete medical examination includes any vaccination. Finance Act 2025 s.6(a)(i) limits that to a vaccine registered with the National Pharmaceutical Regulatory Agency from YA2026."},
 "MEDICAL_VACCINATION":{Act53Scope,"s.46(1)(g)(iii) up to RM1,000 within the RM10,000. For YA2025 the qualifying vaccines are the eight listed in the s.46(1)(g) proviso (d) of the May 2024 text (pneumococcal, HPV, influenza, rotavirus, varicella, meningococcal, TDAP, COVID-19). Finance Act 2025 s.6(a)(ii) replaces that list with NPRA-registered vaccines from YA2026."},
}
