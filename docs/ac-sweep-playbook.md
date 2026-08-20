# AC sweep playbook

Discover every STK, write observable acceptance criteria, stamp acceptance_review, and log acceptance_criteria_witnessed.

Templates: AC-ENFORCE, AC-EMPTY, AC-NEG, AC-DUAL-EQ, AC-OWN.

Concurrency cap: 5 workers.

Exit: every STK is on proof/ac-sweep/checklist.yaml as done, deferred, or not_applicable.
