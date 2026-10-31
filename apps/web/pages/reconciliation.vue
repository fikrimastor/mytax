<script setup lang="ts">
const config=useRuntimeConfig()
const income=ref(70000)
const draft=ref(0)
const pcb=ref(0)
const output=ref<any>()
const busy=ref(false)
const error=ref('')
const money=(sen:number)=>new Intl.NumberFormat('en-MY',{style:'currency',currency:'MYR'}).format(sen/100)
const convert=(n:number)=>{const s=Math.round(n*100);if(!Number.isSafeInteger(s)||s<0)throw new Error('Enter valid non-negative amounts');return s}
async function preview(){
 busy.value=true;error.value='';output.value=null
 try{
  output.value=await $fetch(`${config.public.apiBase}/v1/reconciliation/preview`,{method:'POST',body:{assessment_year:2025,declared_total_income_sen:convert(income.value),draft_reliefs_sen:convert(draft.value),verified_reliefs_sen:0,entered_rebates_sen:0,pcb_sen:convert(pcb.value)}})
 }catch(e:any){error.value=e?.data?.error||e.message||'Request failed'}finally{busy.value=false}
}
</script>
<template>
<main class="container">
 <p class="eyebrow">mytax · Milestone 5 · YA 2025</p>
 <h1>Form BE arithmetic preview</h1>
 <p class="subtle">For demonstration only. Not an official chargeable-income calculation, Form BE reconciliation, tax liability or refund determination.</p>
 <div class="grid">
  <section class="surface">
   <h2>Enter illustrative amounts</h2>
   <form @submit.prevent="preview">
    <div class="field"><label for="income">Declared total income (RM)</label><input id="income" type="number" v-model.number="income" min="0" step="0.01" required /></div>
    <div class="field"><label for="relief">Draft reliefs for review (RM)</label><input id="relief" type="number" v-model.number="draft" min="0" step="0.01" required /><small class="subtle">Not deducted from the computation</small></div>
    <div class="field"><label for="pcb">PCB already deducted (RM)</label><input id="pcb" type="number" v-model.number="pcb" min="0" step="0.01" required /></div>
    <button class="btn" type="submit" :disabled="busy">{{busy?'Calculating…':'Preview worksheet'}}</button>
   </form>
   <p v-if="error" role="alert" style="color:#b91c1c">{{error}}</p>
  </section>
  <section class="surface">
   <h2>Illustrative worksheet</h2>
   <template v-if="output">
    <p class="pill">Not filing-ready</p>
    <p>Declared income: <strong>{{money(output.total_income_sen)}}</strong></p>
    <p>Draft reliefs excluded: <strong>{{money(output.draft_reliefs_excluded_sen)}}</strong></p>
    <p>Hypothetical chargeable income: <strong>{{money(output.illustrative_chargeable_income_sen)}}</strong></p>
    <p>Gross marginal tax: <strong>{{money(output.gross_tax_sen)}}</strong></p>
    <p>PCB input: <strong>{{money(output.pcb_sen)}}</strong></p>
    <hr />
    <h3>Indicative difference: {{money(output.illustrative_balance_sen)}}</h3>
    <p class="subtle">Negative difference does not confirm any refund entitlement.</p>
   </template>
   <p v-else class="subtle">Run a preview to see the arithmetic breakdown.</p>
  </section>
 </div>
 <section v-if="output" class="surface" style="margin-top:20px"><h2>Outstanding checks</h2><ul><li v-for="warn in output.warnings" :key="warn">{{warn}}</li></ul></section>
 <p><NuxtLink to="/assessment">Gross tax brackets</NuxtLink> · <NuxtLink to="/">Claim checker</NuxtLink></p>
</main>
</template>