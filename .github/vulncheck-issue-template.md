---
title: 'Failed build: Security Vulnerability Check'
labels: vulncheck
---

GitHub Actions workflow [{{ env.WORKFLOW }} #{{ env.RUN_NUMBER }}]({{ env.SERVER_URL }}/{{ env.REPOSITORY }}/actions/runs/{{ env.RUN_ID }}) failed.

Event: {{ env.EVENT_NAME }}
Branch: [{{ env.REF_NAME }}]({{ env.SERVER_URL }}/{{ env.REPOSITORY }}/tree/{{ env.REF_NAME }})
Commit: [{{ env.SHA }}]({{ env.SERVER_URL }}/{{ env.REPOSITORY }}/commit/{{ env.SHA }})

<sup><i>Auto-reported by govulncheck</i></sup>
