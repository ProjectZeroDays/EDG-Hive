package server

import "os"

// Legacy ARTEX_* environment support.
//
// The product was renamed from ARTEX to Hive. Existing installs export
// ARTEX_* variables (ARTEX_CONFIG, ARTEX_PG_DSN, ARTEX_SECRET_KEY, ...).
// Rather than forcing every operator to rename their environment at upgrade
// time, this shim mirrors every known legacy variable into its HIVE_*
// equivalent at process start. HIVE_* always wins when both are set.
//
// This file is the only place legacy ARTEX_* names may appear. When the
// legacy support window closes, delete this file.
const legacyCookieName = "artex_token"

var legacyEnv = map[string]string{
	"HIVE_BASE":                 "ARTEX_BASE",
	"HIVE_BTW_MAX_OUTPUT_TOKENS": "ARTEX_BTW_MAX_OUTPUT_TOKENS",
	"HIVE_CONFIG":               "ARTEX_CONFIG",
	"HIVE_DATA":                 "ARTEX_DATA",
	"HIVE_INTENT_BUDGET_TOKENS": "ARTEX_INTENT_BUDGET_TOKENS",
	"HIVE_INTENT_BUDGET_TURNS":  "ARTEX_INTENT_BUDGET_TURNS",
	"HIVE_JWT_KEY":              "ARTEX_JWT_KEY",
	"HIVE_JWT_KEY_DIR":          "ARTEX_JWT_KEY_DIR",
	"HIVE_KEY_DIR":              "ARTEX_KEY_DIR",
	"HIVE_LLM_BASE_URL":         "ARTEX_LLM_BASE_URL",
	"HIVE_LLM_MODEL":            "ARTEX_LLM_MODEL",
	"HIVE_LLM_PROVIDER":         "ARTEX_LLM_PROVIDER",
	"HIVE_LLM_PROXY":            "ARTEX_LLM_PROXY",
	"HIVE_LLM_STREAM":           "ARTEX_LLM_STREAM",
	"HIVE_NOISE_PROFILE":        "ARTEX_NOISE_PROFILE",
	"HIVE_NOTIFY_ALLOW_LOCAL":   "ARTEX_NOTIFY_ALLOW_LOCAL",
	"HIVE_PASS":                 "ARTEX_PASS",
	"HIVE_PG_DSN":               "ARTEX_PG_DSN",
	"HIVE_RENDERED_DENY_PROBE":  "ARTEX_RENDERED_DENY_PROBE",
	"HIVE_REVIEW_LIVE_CONFIG":   "ARTEX_REVIEW_LIVE_CONFIG",
	"HIVE_RSI_LESSONS":          "ARTEX_RSI_LESSONS",
	"HIVE_SCRUB_KEEP":           "ARTEX_SCRUB_KEEP",
	"HIVE_SECRET_KEY":           "ARTEX_SECRET_KEY",
	"HIVE_SKILL_DIR":            "ARTEX_SKILL_DIR",
	"HIVE_TARGET_SEAM_PROBE":    "ARTEX_TARGET_SEAM_PROBE",
	"HIVE_TARGET_USER_AGENT":    "ARTEX_TARGET_USER_AGENT",
	"HIVE_TASK_BUDGET_TOKENS":   "ARTEX_TASK_BUDGET_TOKENS",
	"HIVE_TASK_BUDGET_TURNS":    "ARTEX_TASK_BUDGET_TURNS",
	"HIVE_TOKEN":                "ARTEX_TOKEN",
	"HIVE_TOOL_RAG":             "ARTEX_TOOL_RAG",
	"HIVE_TOOL_RAG_TOPK":        "ARTEX_TOOL_RAG_TOPK",
	"HIVE_TRUSTED_PROXIES":      "ARTEX_TRUSTED_PROXIES",
	"HIVE_UNTRUSTED_BOUNDARY":   "ARTEX_UNTRUSTED_BOUNDARY",
	"HIVE_UPDATE_CHECK":         "ARTEX_UPDATE_CHECK",
	"HIVE_USER":                 "ARTEX_USER",
}

func init() {
	for hive, artex := range legacyEnv {
		if os.Getenv(hive) == "" {
			if v := os.Getenv(artex); v != "" {
				_ = os.Setenv(hive, v)
			}
		}
	}
}
