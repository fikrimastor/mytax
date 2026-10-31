<script setup lang="ts">
const config=useRuntimeConfig()
const income=ref(35000), aggregate=ref(35000), dividends=ref(0)
const resident=ref(false),noBusiness=ref(false),assessmentType=ref('SINGLE')
const data=ref<any>(),error=ref(''),loading=ref(false)
const money=(v:number)=>new Intl.NumberFormat('en-MY',{style:'currency',currency:'MYR'}).format(v/100)
async function run(){
 loading.value=true;error.value='';data.value=null
 try{
  const sen=(n:number)=>{const v=Math.round(n*100);if(!Number.isSafeInteger(v)||v<0)throw Error('Enter non-negative RM amounts');return v}
  data.value=await $fetch(`${config.public.apiBase}/v1/formbe/preflight`,{method:'POST',body:{assessment_year:2025,resident_confirmed:resident.value,no_business_income_confirmed:noBusiness.value,assessment_type:assessmentType.value,employment_income_sen:0,dividend_income_sen:sen(dividends.value),aggregate_income_sen:sen(aggregate.value),chargeable_income_sen:sen(income.value),spouse_relief_verified:false}})
 }catch(e:any){error.value=e?.data?.error||e.message||'API error'}finally{loading.value=false}
}
</script>
<template><main class="container">
 <p class="eyebrow">mytax · Milestone 6</p><h1>Form BE preflight</h1>
 <p class="subtle">YA 2025 developer preview. This is not an approved tax assessment. Do not enter personal identifiers or upload documents.</p>
 <div class="grid">
  <section class="surface"><h2>Assessment facts</h2><form @submit.prevent="run">
   <div class="field"><label for="income">Previously determined chargeable income (RM)</label><input id="income" v-model.number="income" type="number" min="0" step="0.01" required /></div>
   <div class="field"><label for="aggregate">Aggregate income (RM)</label><input id="aggregate" v-model.number="aggregate" type="number" min="0" step="0.01" required /></div>
   <div class="field"><label for="dividends">Dividend income (RM)</label><input id="dividends" v-model.number="dividends" type="number" min="0" step="0.01" required /></div>
   <div class="field"><label for="assessment">Assessment type</label><select id="assessment" v-model="assessmentType"><option value="SINGLE">Single</option><option value="SEPARATE">Separate assessment</option><option value="JOINT">Joint assessment (review required)</option><option value="SPOUSE_NO_INCOME">Spouse without income (review required)</option></select></div>
   <div class="field"><label><input v-model="resident" type="checkbox" style="width:auto"/> Confirm Malaysian tax residence</label></div>
   <div class="field"><label><input v-model="noBusiness" type="checkbox" style="width:auto"/> Confirm no business income</label></div>
   <button class="btn" :disabled="loading">{{loading?'Checking…':'Run preflight'}}</button>
   <p v-if="error" role="alert" style="color:#b91c1c">{{error}}</p>
  </form></section>
  <section class="surface"><h2>Preflight result</h2>
   <template v-if="data">
    <span class="pill">Not filing ready</span>
    <p>Potential self rebate threshold: <strong>{{money(data.indicative_self_rebate_sen)}}</strong></p>
    <p>Actual rebate applied: <strong>{{money(data.rebate_applied_sen)}}</strong></p>
    <div v-for="item in data.checks" :key="item.code" style="padding:9px 0;border-bottom:1px solid #e2e8f0">
     <strong>{{item.code}} — {{item.state}}</strong><p class="subtle" style="margin:4px 0">{{item.message}}</p>
    </div>
   </template>
   <p v-else class="subtle">Preflight checks identify incomplete facts and Form BE calculations. They do not certify eligibility.</p>
  </section>
 </div>
 <p><NuxtLink to="/reconciliation">Reconciliation</NuxtLink> · <NuxtLink to="/">Claim checker</NuxtLink> · <a href="https://www.hasil.gov.my/media/pshpbomm/explanatorynotes_be2025_2.pdf" target="_blank" rel="noopener">Official Form BE 2025 notes</a></p>
</main></template>