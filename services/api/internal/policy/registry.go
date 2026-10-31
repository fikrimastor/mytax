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
  out=append(out,RegistryEntry{ID:e.id,Label:e.label,MaxSen:e.rm*100,ParentID:e.parent,Conditional:true,Note:e.note,SourceURL:OfficialReliefSource,Status:"DRAFT_REVIEW_REQUIRED"})
 }
 return out
}
