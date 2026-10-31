<script setup lang="ts">
const config = useRuntimeConfig()
const assessmentYear = ref(2025)
const amount = ref(129)
const category = ref('LIFESTYLE')
const isDuplicate = ref(false)
const hasEvidence = ref(false)
const result = ref<any>(null)
const error = ref('')
const loading = ref(false)
const check = async () => {
 loading.value = true; error.value = ''; result.value = null
 try {
  result.value = await $fetch(`${config.public.apiBase}/v1/claims/evaluate`, {
   method: 'POST', body: { assessment_year: Number(assessmentYear.value), category: category.value, amount_sen: Math.round(Number(amount.value) * 100), is_duplicate: isDuplicate.value, has_evidence: hasEvidence.value }
  })
 } catch (e:any) { error.value = e?.data?.error || 'Unable to contact the API' }
 finally { loading.value = false }
}
</script>
<template>
<div class="container">
 <header style="display:flex;justify-content:space-between;align-items:center;gap:12px;flex-wrap:wrap">
  <div><div class="eyebrow">Malaysia personal tax · Developer preview</div><h1 style="font-size:38px;margin:8px 0">mytax<span style="color:#0f766e">.</span></h1></div>
  <span class="pill">Rules pending LHDN verification</span>
 </header>
 <p class="subtle" style="max-width:720px">A compliance-first claim checker. This preview performs only preliminary checks and intentionally does not approve deductions or calculate tax before official year-specific policies are reviewed.</p>
 <div class="grid" style="margin-top:28px">
  <section class="surface"><h2>Check a potential claim</h2>
   <form @submit.prevent="check">
    <div class="field"><label>Year of assessment</label><select v-model="assessmentYear"><option :value="2025">YA 2025</option><option :value="2026">YA 2026</option></select></div>
    <div class="field"><label>Category (illustrative)</label><select v-model="category"><option value="LIFESTYLE">Lifestyle</option><option value="MEDICAL">Medical</option><option value="EDUCATION">Education</option><option value="OTHER">Other</option></select></div>
    <div class="field"><label>Amount (RM)</label><input v-model.number="amount" type="number" min="0.01" step="0.01" required /></div>
    <div class="field"><label><input type="checkbox" v-model="hasEvidence" style="width:auto" /> I have supporting evidence</label></div>
    <div class="field"><label><input type="checkbox" v-model="isDuplicate" style="width:auto" /> Already claimed this expense elsewhere</label></div>
    <button class="btn" type="submit" :disabled="loading">{{ loading ? 'Checking…' : 'Run preliminary check' }}</button>
   </form>
  </section>
  <section class="surface"><h2>Assessment outcome</h2><p v-if="!result && !error" class="subtle">Complete the form to evaluate basic evidence and duplicate flags.</p><p v-if="error" style="color:#b91c1c">{{ error }}</p>
    <div v-if="result"><p class="eyebrow">{{ result.status }}</p><h3>Eligible amount: RM {{ (result.eligible_amount_sen / 100).toFixed(2) }}</h3><p class="subtle">{{ result.disclaimer }}</p><strong>Why?</strong><ul><li v-for="r in result.reasons" :key="r">{{ r.replaceAll('_',' ') }}</li></ul></div>
    <hr style="border:0;border-top:1px solid #e2e8f0;margin:30px 0" /><h3>Safety by design</h3><p class="subtle">OCR is evidence extraction, not legal validation. Approved rule packs, eligibility conditions and official examples must be installed before production calculations are enabled.</p>
  </section>
 </div>
 <nav style="margin-top:24px"><NuxtLink to="/reconciliation">Form BE worksheet preview</NuxtLink> · <NuxtLink to="/assessment">Gross tax preview</NuxtLink></nav>
 <footer class="subtle" style="padding:36px 0;font-size:13px">Not affiliated with LHDN. Do not enter sensitive personal tax details in this prototype.</footer>
</div>
</template>