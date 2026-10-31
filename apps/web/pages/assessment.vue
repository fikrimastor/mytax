<script setup lang="ts">
const config=useRuntimeConfig()
const ya=ref(2025)
const incomeRM=ref(70000)
const estimate=ref<any>(null)
const busy=ref(false)
const error=ref('')
const format=(sen:number)=>new Intl.NumberFormat('en-MY',{style:'currency',currency:'MYR'}).format(sen/100)
const run=async()=>{
 busy.value=true;error.value='';estimate.value=null
 try{
  const sen=Math.round(Number(incomeRM.value)*100)
  if(!Number.isSafeInteger(sen)||sen<0)throw new Error('Enter a valid non-negative amount')
  estimate.value=await $fetch(`${config.public.apiBase}/v1/assessment/preview`,{method:'POST',body:{assessment_year:ya.value,declared_chargeable_income_sen:sen,claims:[]}})
 }catch(e:any){error.value=e?.data?.error||e?.message||'Request failed'}
 finally{busy.value=false}
}
</script>
<template>
<main class="container">
 <p class="eyebrow">mytax · Milestone 4</p>
 <h1>Assessment preview</h1>
 <p class="subtle">Illustrative YA 2025 resident gross tax on <strong>chargeable income you already determined</strong>. This is not taxable salary, a refund prediction, or a completed Form BE calculation.</p>
 <div class="grid">
  <section class="surface">
   <h2>Calculate gross tax</h2>
   <form @submit.prevent="run">
    <div class="field"><label for="year">Assessment year</label><select id="year" v-model.number="ya"><option :value="2025">YA 2025</option><option :value="2026">YA 2026 (not verified)</option></select></div>
    <div class="field"><label for="chargeable">Declared chargeable income (RM)</label><input id="chargeable" v-model.number="incomeRM" type="number" min="0" step="0.01" required /></div>
    <button class="btn" :disabled="busy">{{busy?'Calculating…':'Preview assessment'}}</button>
   </form>
   <p v-if="error" role="alert" style="color:#b91c1c">{{error}}</p>
  </section>
  <section class="surface">
   <h2>Gross tax estimate</h2>
   <template v-if="estimate">
    <div class="eyebrow">{{estimate.status}}</div>
    <h2 style="font-size:34px">{{format(estimate.gross_tax.gross_tax_sen)}}</h2>
    <p class="subtle">Not net tax payable or refundable.</p>
    <h3>Bracket breakdown</h3>
    <div v-for="line in estimate.gross_tax.lines" :key="line.from_sen" style="display:flex;justify-content:space-between;border-bottom:1px solid #e2e8f0;padding:8px 0;gap:8px">
     <span>{{format(line.from_sen)}} – {{format(line.to_sen)}} ({{line.rate_percent}}%)</span><strong>{{format(line.tax_sen)}}</strong>
    </div>
    <p><strong>Draft reliefs deducted:</strong> RM0.00</p>
    <p><strong>Reconciliation:</strong> {{estimate.reconciliation.invariant_satisfied?'Passed':'FAILED'}}</p>
    <p class="pill">Filing ready: No</p>
   </template>
   <p v-else class="subtle">Enter an amount to view a marginal-rate breakdown.</p>
  </section>
 </div>
 <section class="surface" style="margin-top:20px">
  <h2>Before filing with LHDN</h2>
  <p class="subtle">This preview does not establish chargeable income, validate relief entitlement, apply rebates, reconcile PCB or implement the exact official Form BE computation. No draft reliefs are included in the tax computation.</p>
  <ul v-if="estimate"><li v-for="warning in estimate.warnings" :key="warning">{{warning}}</li></ul>
  <NuxtLink to="/">Return to claim checker</NuxtLink>
  <p><a href="https://www.hasil.gov.my/individu/kadar-cukai/" target="_blank" rel="noopener noreferrer">Official LHDN tax rate source</a></p>
 </section>
</main>
</template>