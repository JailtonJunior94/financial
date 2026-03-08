#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <task-execution-report.md>"
  exit 2
fi

report_file="$1"

if [[ ! -f "$report_file" ]]; then
  echo "ERROR: report file not found: $report_file"
  exit 2
fi

missing=0

require_pattern() {
  local pattern="$1"
  local label="$2"

  if ! grep -Eiq "$pattern" "$report_file"; then
    echo "MISSING: $label"
    missing=1
  fi
}

# Required by evidence-policy.md
require_pattern "executed commands|comandos executados" "Executed commands section"
require_pattern "changed files|arquivos alterados" "Changed files section"
require_pattern "validation results|resultado(s)? de validacao" "Validation results section"
require_pattern "assumptions|premissas" "Assumptions section"
require_pattern "residual risks|riscos residuais" "Residual risks section"

# Required by executar-task.md closing gate
require_pattern "test(s)?[[:space:]]*:" "Test evidence"
require_pattern "lint[[:space:]]*:" "Lint evidence"
require_pattern "code-reviewer|technical review" "Code-reviewer evidence"
require_pattern "qa[[:space:]]*report|qa[[:space:]]*:" "QA evidence"

if [[ $missing -ne 0 ]]; then
  echo ""
  echo "Evidence bundle validation failed: $report_file"
  exit 1
fi

echo "Evidence bundle validation passed: $report_file"
