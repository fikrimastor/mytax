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
  if diff,ok:=act53Differences[e.id];ok{note+=" | ACT53_DIFFERS: "+diff}
  out=append(out,RegistryEntry{ID:e.id,Label:e.label,MaxSen:e.rm*100,ParentID:e.parent,Conditional:true,Note:note,SourceURL:OfficialReliefSource,Status:"DRAFT_REVIEW_REQUIRED"})
 }
 return out
}

// act53Differences lists registry entries whose amount or scope could NOT be confirmed
// against Income Tax Act 1967 (Act 53) updated text as at 21 May 2024 (not an authentic
// text; later amendments not checked). MaxSen stays the LHDN YA2025 table value; do not
// treat any of these as confirmed. See docs/ACT53_RECONCILIATION_YA2025.md.
var act53Differences=map[string]string{
 "DISABLED_SELF":"Act s.46(1)(e) gives a further RM6,000; LHDN YA2025 table gives RM7,000. Unresolved.",
 "MEDICAL_LEARNING_DISABILITY":"Act s.46(1)(ha) gives RM4,000 (part of RM10,000); LHDN YA2025 table gives RM6,000. Unresolved.",
 "CHILDCARE":"Act s.46(1)(r) gives RM2,000 plus a further RM1,000 only for YA2020-2024; LHDN YA2025 table gives RM3,000. Unresolved.",
 "SSPN":"Act s.46(1)(k) gives RM8,000 for his child; a footnote limits that amendment to YA2019-2024; YA2025 basis not in this text. Unresolved.",
 "EDU_MED_INSURANCE":"Act s.49(1B) gives RM3,000; LHDN YA2025 table gives RM4,000. Unresolved.",
 "DISABLED_SPOUSE":"Act s.47(1)(b) gives a further RM5,000; LHDN YA2025 table gives RM6,000. Unresolved.",
 "DISABLED_CHILD":"Act s.48(2)(b) gives RM6,000; LHDN YA2025 table gives RM8,000. Unresolved.",
 "EV_COMPOST":"Act s.46(1)(v) covers EV charging only (YA2023-2027); composter not found in the text.",
 "MEDICAL_SCREENING":"Act s.46(1)(h) covers complete medical exam, COVID-19 test and mental health only; self-test equipment and disease-detection test fees in the LHDN table were not found.",
 "FIRST_HOME_INTEREST_LOW":"No matching provision in the Act text as at 21 May 2024 (s.46B covers only SPA dated 10 Mar 2009 to 31 Dec 2010). Source for YA2025 unresolved.",
 "FIRST_HOME_INTEREST_MID":"No matching provision in the Act text as at 21 May 2024 (s.46B covers only SPA dated 10 Mar 2009 to 31 Dec 2010). Source for YA2025 unresolved.",
}
