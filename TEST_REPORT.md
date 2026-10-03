?   	github.com/adlanegrm-oss/einvoice-saas	[no test files]
?   	github.com/adlanegrm-oss/einvoice-saas/cmd/outbox_worker	[no test files]
=== RUN   TestHealthAndStatic
--- PASS: TestHealthAndStatic (1.78s)
=== RUN   TestAPIRequiresAuthentication
2026/10/03 20:05:08 INFO rapport journalier g├®n├®r├® date=2026-10-03 factures=0 total_ttc=0
--- PASS: TestAPIRequiresAuthentication (2.45s)
=== RUN   TestLogin
2026/10/03 20:05:10 WARN ├®chec de connexion ip=127.0.0.1
2026/10/03 20:05:10 WARN ├®chec de connexion ip=127.0.0.1
--- PASS: TestLogin (2.83s)
=== RUN   TestPasswordResetFlow
2026/10/03 20:05:13 WARN ├®chec de connexion ip=127.0.0.1
--- PASS: TestPasswordResetFlow (2.90s)
=== RUN   TestDepositListDownloadIsolation
2026/10/03 20:05:16 INFO document archiv├® tenant=t-f93fa2e5fb592009 stored_as=20261003-190516-f32ee564-facture.xml status=VALIDE_PRET_A_ENVOYER format=UBL
2026/10/03 20:05:16 INFO document archiv├® tenant=t-f93fa2e5fb592009 stored_as=20261003-190516-0707068f-brouillon.xml status=BROUILLON_TEMPORAIRE_72H format=UNKNOWN
--- PASS: TestDepositListDownloadIsolation (2.52s)
=== RUN   TestStructuredInvoices
--- PASS: TestStructuredInvoices (2.44s)
=== RUN   TestPurgeExpiredDrafts
2026/10/03 20:05:20 INFO document archiv├® tenant=t-f93fa2e5fb592009 stored_as=20261003-190520-132785fe-vieux.xml status=BROUILLON_TEMPORAIRE_72H format=UNKNOWN
2026/10/03 20:05:20 INFO document archiv├® tenant=t-f93fa2e5fb592009 stored_as=20261003-190520-33a760a3-recent.xml status=BROUILLON_TEMPORAIRE_72H format=UNKNOWN
--- PASS: TestPurgeExpiredDrafts (1.87s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/cmd/server	16.888s
?   	github.com/adlanegrm-oss/einvoice-saas/internal	[no test files]
=== RUN   TestAuditChainIntegrityAndTampering
    chain_test.go:33: Succ├¿s : Alt├®ration intercept├®e avec succ├¿s -> alt├®ration payload sur evt evt_2_inv_2026_001 (seq 2): hash stock├® 71a77929df4ce9827132ffb274f67e854be857af7ee2fdbb1a6cef31bf6315ef, recalcul├® a0551bca330baa2bbfd272aa35b4fea5bb8d5527783194460771c924384b557f
--- PASS: TestAuditChainIntegrityAndTampering (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/audit	(cached)
=== RUN   TestPasswordHashAndVerify
--- PASS: TestPasswordHashAndVerify (1.33s)
=== RUN   TestPasswordPolicy
--- PASS: TestPasswordPolicy (0.00s)
=== RUN   TestTokenIssueVerify
--- PASS: TestTokenIssueVerify (0.00s)
=== RUN   TestTokenRejections
--- PASS: TestTokenRejections (0.00s)
=== RUN   TestSecretTooShort
--- PASS: TestSecretTooShort (0.00s)
=== RUN   TestStoreAuthenticate
--- PASS: TestStoreAuthenticate (1.96s)
=== RUN   TestResetFlow
--- PASS: TestResetFlow (1.98s)
=== RUN   TestResetTokenExpiryAndReplacement
--- PASS: TestResetTokenExpiryAndReplacement (1.63s)
=== RUN   TestTenantIDStable
--- PASS: TestTenantIDStable (0.00s)
=== RUN   TestLimiter
--- PASS: TestLimiter (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/auth	6.971s
=== RUN   TestPPFConnector_SubmitAndCheck
--- PASS: TestPPFConnector_SubmitAndCheck (0.00s)
=== RUN   TestKSeFConnector_SubmitAndUPO
--- PASS: TestKSeFConnector_SubmitAndUPO (0.00s)
=== RUN   TestDefaultService_ReturnsNotConfigured
--- PASS: TestDefaultService_ReturnsNotConfigured (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/clearance	0.080s
=== RUN   TestValidator_EN16931
--- PASS: TestValidator_EN16931 (0.00s)
=== RUN   TestParseSVRL_WithFailures
--- PASS: TestParseSVRL_WithFailures (0.00s)
=== RUN   TestQuickValidateProfile_Success
--- PASS: TestQuickValidateProfile_Success (0.00s)
=== RUN   TestQuickValidateProfile_MissingMandatoryFields
--- PASS: TestQuickValidateProfile_MissingMandatoryFields (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/compliance/en16931	0.089s
=== RUN   TestLoadDevDefaults
--- PASS: TestLoadDevDefaults (0.00s)
=== RUN   TestLoadProdRequiresSecrets
--- PASS: TestLoadProdRequiresSecrets (0.00s)
=== RUN   TestLoadRejectsBadValues
--- PASS: TestLoadRejectsBadValues (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/config	0.073s
=== RUN   TestPrecisionAndRounding
--- PASS: TestPrecisionAndRounding (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/currency	(cached)
=== RUN   TestMonetaryCalculations
--- PASS: TestMonetaryCalculations (0.00s)
=== RUN   TestExternalStateProjection
--- PASS: TestExternalStateProjection (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/domain	(cached)
=== RUN   TestInvoiceModel_StructureAndAmounts
=== RUN   TestInvoiceModel_StructureAndAmounts/Contr├┤le_coh├®rence_arithm├®tique_montants
=== RUN   TestInvoiceModel_StructureAndAmounts/Pr├®sence_identifiants_obligatoires
--- PASS: TestInvoiceModel_StructureAndAmounts (0.00s)
    --- PASS: TestInvoiceModel_StructureAndAmounts/Contr├┤le_coh├®rence_arithm├®tique_montants (0.00s)
    --- PASS: TestInvoiceModel_StructureAndAmounts/Pr├®sence_identifiants_obligatoires (0.00s)
=== RUN   TestMoney_Operations
=== RUN   TestMoney_Operations/Addition_sans_d├®rive_binaire_de_flottant
=== RUN   TestMoney_Operations/Alignement_des_├®chelles_(Prix_unitaire_├á_4_d├®cimales_+_Totaux)
=== RUN   TestMoney_Operations/Rejet_addition_devises_diff├®rentes
--- PASS: TestMoney_Operations (0.00s)
    --- PASS: TestMoney_Operations/Addition_sans_d├®rive_binaire_de_flottant (0.00s)
    --- PASS: TestMoney_Operations/Alignement_des_├®chelles_(Prix_unitaire_├á_4_d├®cimales_+_Totaux) (0.00s)
    --- PASS: TestMoney_Operations/Rejet_addition_devises_diff├®rentes (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/domain/models	(cached)
=== RUN   TestGenerateFacturXXML
--- PASS: TestGenerateFacturXXML (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/exporter	(cached)
?   	github.com/adlanegrm-oss/einvoice-saas/internal/exporter/facturae	[no test files]
?   	github.com/adlanegrm-oss/einvoice-saas/internal/exporter/fatturapa	[no test files]
?   	github.com/adlanegrm-oss/einvoice-saas/internal/exporter/ksef	[no test files]
=== RUN   TestPipelineHandler_EmitInvoice_SuccessAndAudit
--- PASS: TestPipelineHandler_EmitInvoice_SuccessAndAudit (0.06s)
=== RUN   TestPipelineHandler_EmitInvoice_ValidationError
--- PASS: TestPipelineHandler_EmitInvoice_ValidationError (0.00s)
=== RUN   TestIntegration_Deposit_MultiTenant_And_Purge
=== RUN   TestIntegration_Deposit_MultiTenant_And_Purge/D├®p├┤t_conforme_Tenant-Alpha_->_ACCEPTE
2026/10/03 20:05:04 INFO document archiv├® tenant=tenant-alpha stored_as=20261003-190504-49cb845a-facture_alpha.xml status=VALIDE_PRET_A_ENVOYER format=UBL
=== RUN   TestIntegration_Deposit_MultiTenant_And_Purge/Isolation_Multi-Tenant_:_Tenant-Beta_ne_voit_pas_les_documents_de_Tenant-Alpha
=== RUN   TestIntegration_Deposit_MultiTenant_And_Purge/Isolation_Multi-Tenant_:_Tentative_de_t├®l├®chargement_direct_inter-tenant_->_404
=== RUN   TestIntegration_Deposit_MultiTenant_And_Purge/D├®p├┤t_Draft_(Brouillon_temporaire_72h)_et_cycle_de_purge
2026/10/03 20:05:04 INFO document archiv├® tenant=tenant-alpha stored_as=20261003-190504-7020bd9e-draft_invoice.xml status=BROUILLON_TEMPORAIRE_72H format=UBL
--- PASS: TestIntegration_Deposit_MultiTenant_And_Purge (0.01s)
    --- PASS: TestIntegration_Deposit_MultiTenant_And_Purge/D├®p├┤t_conforme_Tenant-Alpha_->_ACCEPTE (0.01s)
    --- PASS: TestIntegration_Deposit_MultiTenant_And_Purge/Isolation_Multi-Tenant_:_Tenant-Beta_ne_voit_pas_les_documents_de_Tenant-Alpha (0.00s)
    --- PASS: TestIntegration_Deposit_MultiTenant_And_Purge/Isolation_Multi-Tenant_:_Tentative_de_t├®l├®chargement_direct_inter-tenant_->_404 (0.00s)
    --- PASS: TestIntegration_Deposit_MultiTenant_And_Purge/D├®p├┤t_Draft_(Brouillon_temporaire_72h)_et_cycle_de_purge (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/handler	0.162s
=== RUN   TestIdempotencyStore_TenantIsolation
--- PASS: TestIdempotencyStore_TenantIsolation (0.05s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/idempotency	0.135s
=== RUN   TestValidateInvoice
=== RUN   TestValidateInvoice/Facture_valide_avec_calcul_de_totaux
=== RUN   TestValidateInvoice/Num├â┬®ro_de_facture_manquant
=== RUN   TestValidateInvoice/Aucun_article_dans_la_facture
=== RUN   TestValidateInvoice/Quantit├â┬®_invalide_(<=_0)
--- PASS: TestValidateInvoice (0.00s)
    --- PASS: TestValidateInvoice/Facture_valide_avec_calcul_de_totaux (0.00s)
    --- PASS: TestValidateInvoice/Num├â┬®ro_de_facture_manquant (0.00s)
    --- PASS: TestValidateInvoice/Aucun_article_dans_la_facture (0.00s)
    --- PASS: TestValidateInvoice/Quantit├â┬®_invalide_(<=_0) (0.00s)
=== RUN   TestCalculateTotals
--- PASS: TestCalculateTotals (0.00s)
=== RUN   TestCalculateTotalsRounding
--- PASS: TestCalculateTotalsRounding (0.00s)
=== RUN   TestValidateRejectsAbnormalValues
=== RUN   TestValidateRejectsAbnormalValues/TVA_n├â┬®gative
=== RUN   TestValidateRejectsAbnormalValues/TVA_sup├â┬®rieure_100
=== RUN   TestValidateRejectsAbnormalValues/prix_n├â┬®gatif
=== RUN   TestValidateRejectsAbnormalValues/num├â┬®ro_trop_long
=== RUN   TestValidateRejectsAbnormalValues/num├â┬®ro_d'espaces
--- PASS: TestValidateRejectsAbnormalValues (0.00s)
    --- PASS: TestValidateRejectsAbnormalValues/TVA_n├â┬®gative (0.00s)
    --- PASS: TestValidateRejectsAbnormalValues/TVA_sup├â┬®rieure_100 (0.00s)
    --- PASS: TestValidateRejectsAbnormalValues/prix_n├â┬®gatif (0.00s)
    --- PASS: TestValidateRejectsAbnormalValues/num├â┬®ro_trop_long (0.00s)
    --- PASS: TestValidateRejectsAbnormalValues/num├â┬®ro_d'espaces (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/invoice	(cached)
=== RUN   TestTracker_LifecycleAndConcurrency
=== RUN   TestTracker_LifecycleAndConcurrency/Filtrage_des_t├óches_actives
=== RUN   TestTracker_LifecycleAndConcurrency/Acc├¿s_concurrentiel_s├®curis├®_(Data_Race_check)
--- PASS: TestTracker_LifecycleAndConcurrency (0.00s)
    --- PASS: TestTracker_LifecycleAndConcurrency/Filtrage_des_t├óches_actives (0.00s)
    --- PASS: TestTracker_LifecycleAndConcurrency/Acc├¿s_concurrentiel_s├®curis├®_(Data_Race_check) (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/jobs	(cached)
=== RUN   TestStateMachine_Transitions
--- PASS: TestStateMachine_Transitions (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/lifecycle/status	(cached)
?   	github.com/adlanegrm-oss/einvoice-saas/internal/logger	[no test files]
=== RUN   TestAuthenticate
=== RUN   TestAuthenticate/sans_jeton_->_401
=== RUN   TestAuthenticate/en-t├¬te_X-User-Role_forg├®_->_401
=== RUN   TestAuthenticate/Authorization:_ADMIN_(ancien_sch├®ma)_->_401
=== RUN   TestAuthenticate/jeton_bidon_->_401
=== RUN   TestAuthenticate/jeton_client_sur_route_partag├®e_->_200
=== RUN   TestAuthenticate/jeton_client_sur_route_admin_->_403
=== RUN   TestAuthenticate/jeton_admin_sur_route_admin_->_200
=== RUN   TestAuthenticate/pr├®fixe_bearer_insensible_├á_la_casse_->_200
--- PASS: TestAuthenticate (0.00s)
    --- PASS: TestAuthenticate/sans_jeton_->_401 (0.00s)
    --- PASS: TestAuthenticate/en-t├¬te_X-User-Role_forg├®_->_401 (0.00s)
    --- PASS: TestAuthenticate/Authorization:_ADMIN_(ancien_sch├®ma)_->_401 (0.00s)
    --- PASS: TestAuthenticate/jeton_bidon_->_401 (0.00s)
    --- PASS: TestAuthenticate/jeton_client_sur_route_partag├®e_->_200 (0.00s)
    --- PASS: TestAuthenticate/jeton_client_sur_route_admin_->_403 (0.00s)
    --- PASS: TestAuthenticate/jeton_admin_sur_route_admin_->_200 (0.00s)
    --- PASS: TestAuthenticate/pr├®fixe_bearer_insensible_├á_la_casse_->_200 (0.00s)
=== RUN   TestExpiredToken
--- PASS: TestExpiredToken (1.10s)
=== RUN   TestClaimsInContext
--- PASS: TestClaimsInContext (0.00s)
=== RUN   TestRateLimit
--- PASS: TestRateLimit (0.00s)
=== RUN   TestRecover
2026/10/03 20:05:05 ERROR panique dans un handler path=/test panic=boom stack="goroutine 25 [running]:\nruntime/debug.Stack()\n\tC:/Program Files/Go/src/runtime/debug/stack.go:26 +0x5e\ngithub.com/adlanegrm-oss/einvoice-saas/internal/middleware.Recover.func1.1()\n\tD:/einvoice-saas-release1/einvoice-saas/internal/middleware/auth.go:138 +0x58\npanic({0x7ff6df4c6980?, 0x7ff6deffed50?})\n\tC:/Program Files/Go/src/runtime/panic.go:859 +0x125\ngithub.com/adlanegrm-oss/einvoice-saas/internal/middleware.TestRecover.func1({0x1a06b6c0300?, 0x24c86fd3618?}, 0x1a06b69aab0?)\n\tD:/einvoice-saas-release1/einvoice-saas/internal/middleware/auth_test.go:102 +0x25\nnet/http.HandlerFunc.ServeHTTP(0x7ff6deb6ba6f?, {0x7ff6df538bf0?, 0x1a06b6a23c0?}, 0x7ff6deb0439d?)\n\tC:/Program Files/Go/src/net/http/server.go:2338 +0x29\ngithub.com/adlanegrm-oss/einvoice-saas/internal/middleware.Recover.func1({0x7ff6df538bf0?, 0x1a06b6a23c0?}, 0x1a06b580008?)\n\tD:/einvoice-saas-release1/einvoice-saas/internal/middleware/auth.go:142 +0x6c\nnet/http.HandlerFunc.ServeHTTP(0x38?, {0x7ff6df538bf0?, 0x1a06b6a23c0?}, 0x7ff6df5409a0?)\n\tC:/Program Files/Go/src/net/http/server.go:2338 +0x29\ngithub.com/adlanegrm-oss/einvoice-saas/internal/middleware.do({0x7ff6df536260, 0x1a06b6982a0}, 0x0)\n\tD:/einvoice-saas-release1/einvoice-saas/internal/middleware/auth_test.go:31 +0x256\ngithub.com/adlanegrm-oss/einvoice-saas/internal/middleware.TestRecover(0x1a06b6de400)\n\tD:/einvoice-saas-release1/einvoice-saas/internal/middleware/auth_test.go:103 +0x5b\ntesting.tRunner(0x1a06b6de400, 0x7ff6df53d6a8)\n\tC:/Program Files/Go/src/testing/testing.go:2193 +0xc3\ncreated by testing.(*T).Run in goroutine 1\n\tC:/Program Files/Go/src/testing/testing.go:2258 +0x4b4\n"
--- PASS: TestRecover (0.03s)
=== RUN   TestRateLimiter_TokenBucket_IP
--- PASS: TestRateLimiter_TokenBucket_IP (0.00s)
=== RUN   TestRateLimiter_AccountProtection
--- PASS: TestRateLimiter_AccountProtection (0.00s)
=== RUN   TestLogSanitizer_MasksSecrets
--- PASS: TestLogSanitizer_MasksSecrets (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/middleware	1.222s
=== RUN   TestMetricsAndTraceMiddleware
--- PASS: TestMetricsAndTraceMiddleware (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/observability	0.057s
=== RUN   TestParse_EDIFACT
=== RUN   TestParse_EDIFACT/Message_EDIFACT_valide_avec_BGM_et_DTM
=== RUN   TestParse_EDIFACT/Message_sans_num├®ro_BGM_utilise_le_fallback_UNKNOWN
=== RUN   TestParse_EDIFACT/Erreur_sur_document_vide
--- PASS: TestParse_EDIFACT (0.00s)
    --- PASS: TestParse_EDIFACT/Message_EDIFACT_valide_avec_BGM_et_DTM (0.00s)
    --- PASS: TestParse_EDIFACT/Message_sans_num├®ro_BGM_utilise_le_fallback_UNKNOWN (0.00s)
    --- PASS: TestParse_EDIFACT/Erreur_sur_document_vide (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/parser/edifact	(cached)
?   	github.com/adlanegrm-oss/einvoice-saas/internal/parser/facturae	[no test files]
?   	github.com/adlanegrm-oss/einvoice-saas/internal/parser/fatturapa	[no test files]
?   	github.com/adlanegrm-oss/einvoice-saas/internal/parser/ksef	[no test files]
=== RUN   TestParse_UBL
=== RUN   TestParse_UBL/Facture_UBL_valide_avec_PartyName
=== RUN   TestParse_UBL/Repli_sur_PartyLegalEntity/RegistrationName_si_PartyName_absent_(EN_16931)
=== RUN   TestParse_UBL/Erreur_sur_XML_malform├®
--- PASS: TestParse_UBL (0.00s)
    --- PASS: TestParse_UBL/Facture_UBL_valide_avec_PartyName (0.00s)
    --- PASS: TestParse_UBL/Repli_sur_PartyLegalEntity/RegistrationName_si_PartyName_absent_(EN_16931) (0.00s)
    --- PASS: TestParse_UBL/Erreur_sur_XML_malform├® (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/parser/ubl	(cached)
=== RUN   TestGenerateInvoiceHash
--- PASS: TestGenerateInvoiceHash (0.00s)
=== RUN   TestSQLiteInvoiceRepository
--- PASS: TestSQLiteInvoiceRepository (0.00s)
=== RUN   TestOwnerIsolation
--- PASS: TestOwnerIsolation (0.00s)
=== RUN   TestDuplicates
--- PASS: TestDuplicates (0.00s)
=== RUN   TestDailyReport
--- PASS: TestDailyReport (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/repository	(cached)
=== RUN   TestRetentionCutoff
--- PASS: TestRetentionCutoff (0.00s)
=== RUN   TestPurgeCycle_UnclosedGuard
--- PASS: TestPurgeCycle_UnclosedGuard (0.00s)
=== RUN   TestPurgeCycle_ExecutionSuccess
2026/10/03 03:41:22 INFO purge cycle achev├®e avec succ├¿s period=2026-06 rows_db=1
--- PASS: TestPurgeCycle_ExecutionSuccess (0.05s)
=== RUN   TestPurgeCycle_Idempotence
2026/10/03 03:41:22 INFO p├®riode d├®j├á purg├®e, op├®ration ignor├®e period=2026-05
--- PASS: TestPurgeCycle_Idempotence (0.00s)
=== RUN   TestPurgeCycle_CrashRecoveryDBPurged
2026/10/03 03:41:22 WARN reprise de purge : DB d├®j├á nettoy├®e, tentative filesystem period=2026-05
2026/10/03 03:41:22 INFO purge cycle achev├®e avec succ├¿s period=2026-05 rows_db=0
--- PASS: TestPurgeCycle_CrashRecoveryDBPurged (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/retention	(cached)
=== RUN   TestAS4Client_SendPayload
--- PASS: TestAS4Client_SendPayload (0.01s)
=== RUN   TestAS4Client_SendPayload_Simulated
--- PASS: TestAS4Client_SendPayload_Simulated (0.00s)
=== RUN   TestOutboxWorker_SuccessFlow
--- PASS: TestOutboxWorker_SuccessFlow (0.00s)
=== RUN   TestOutboxWorker_RetryAndDLQ
--- PASS: TestOutboxWorker_RetryAndDLQ (0.09s)
=== RUN   TestOutboxWorker_BackoffCalculation
--- PASS: TestOutboxWorker_BackoffCalculation (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/routing/as4	(cached)
=== RUN   TestCachedDirectory_Lookup
--- PASS: TestCachedDirectory_Lookup (0.01s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/routing/directory	(cached)
=== RUN   TestDispatcher_RouteAndDispatch
--- PASS: TestDispatcher_RouteAndDispatch (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/routing/dispatcher	(cached)
=== RUN   TestMultiTenantEmissionPipeline
--- PASS: TestMultiTenantEmissionPipeline (0.05s)
=== RUN   TestCrossBorder_FatturaPA_To_KSeF
--- PASS: TestCrossBorder_FatturaPA_To_KSeF (0.00s)
=== RUN   TestPipeline_ConformeEtAcheminee
--- PASS: TestPipeline_ConformeEtAcheminee (0.00s)
=== RUN   TestPipeline_RejetReglementaire
--- PASS: TestPipeline_RejetReglementaire (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/service	0.103s
?   	github.com/adlanegrm-oss/einvoice-saas/internal/tenant	[no test files]
=== RUN   TestEmptyAndUnknown
--- PASS: TestEmptyAndUnknown (0.00s)
=== RUN   TestUBL
--- PASS: TestUBL (0.00s)
=== RUN   TestUBLLegalEntityOnly
--- PASS: TestUBLLegalEntityOnly (0.00s)
=== RUN   TestXMLRejections
--- PASS: TestXMLRejections (0.00s)
=== RUN   TestCII
--- PASS: TestCII (0.00s)
=== RUN   TestEDIFACT
--- PASS: TestEDIFACT (0.00s)
=== RUN   TestPDF
--- PASS: TestPDF (0.00s)
=== RUN   TestSamplePDF
--- PASS: TestSamplePDF (0.00s)
=== RUN   TestEN16931Validation
--- PASS: TestEN16931Validation (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/validator	(cached)
=== RUN   TestPoolRunsJobs
--- PASS: TestPoolRunsJobs (0.00s)
=== RUN   TestSubmitAfterStopDoesNotPanic
--- PASS: TestSubmitAfterStopDoesNotPanic (0.00s)
=== RUN   TestPanickingJobDoesNotKillWorker
2026/10/03 20:05:04 [Worker 1] panique dans un job : boom
--- PASS: TestPanickingJobDoesNotKillWorker (0.05s)
=== RUN   TestFullQueueRejects
2026/10/03 20:05:04 [Worker Pool] File d'attente pleine, t├óche rejet├®e.
--- PASS: TestFullQueueRejects (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/internal/worker	0.088s
=== RUN   TestE2E_FullPipeline_Sprint3
--- PASS: TestE2E_FullPipeline_Sprint3 (0.00s)
PASS
ok  	github.com/adlanegrm-oss/einvoice-saas/test	0.042s
