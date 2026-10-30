package policy

// YA2025Draft is a deliberately INACTIVE subset of the official relief schedule.
// Group allocations represent accounting limits only; they do not prove entitlement.
// No production tax computation may depend on this draft catalog.
func YA2025Draft()Package{
 src:=[]Source{{URL:"https://www.hasil.gov.my/individu/pelepasan-cukai/",Section:"Tahun Taksiran 2025",RetrievedOn:"2026-10-31"}}
 return Package{AssessmentYear:2025,Version:"ya2025-draft-001",Status:Draft,
 Groups:[]Group{{ID:"LIFESTYLE",LimitSen:250000},{ID:"MEDICAL_SELF_FAMILY",LimitSen:1000000},{ID:"EDUCATION_SELF",LimitSen:700000}},
 Rules:[]Rule{
  {ID:"LIFESTYLE_INTERNET",Group:"LIFESTYLE",LimitSen:250000,RequiredFacts:[]string{"resident","within_year","internet_in_own_name","qualifying_expense","no_other_claim"},Sources:src},
  {ID:"LIFESTYLE_BOOKS",Group:"LIFESTYLE",LimitSen:250000,RequiredFacts:[]string{"resident","within_year","qualifying_reading_material","qualifying_beneficiary","no_other_claim"},Sources:src},
  {ID:"MEDICAL_VACCINATION",Group:"MEDICAL_SELF_FAMILY",LimitSen:100000,RequiredFacts:[]string{"resident","within_year","qualifying_vaccine","qualifying_beneficiary","no_other_claim"},Sources:src},
  {ID:"EDUCATION_SELF_SKILLS",Group:"EDUCATION_SELF",LimitSen:200000,RequiredFacts:[]string{"resident","within_year","qualifying_course","own_education","no_other_claim"},Sources:src},
  {ID:"EDUCATION_SELF_OTHER",Group:"EDUCATION_SELF",LimitSen:700000,RequiredFacts:[]string{"resident","within_year","qualifying_course","own_education","no_other_claim"},Sources:src},
 }}
}
