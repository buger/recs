# Unbiased surface-close loop

Cold extract, dual-surface-first GAP queue, then STK ACs or honest residual.

```
proof surfaces extract --unbiased --package crm --out proof/surfaces/runs/crm-cold
proof surfaces gap-queue --unbiased --package crm --format json
proof audit --check surface_coverage
proof audit --check acceptance_criteria_witnessed --fail-level warn
```

Do not use seeds or prior residual notes as disposition authority on the cold run.
