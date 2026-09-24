# Risks

- **PocketBase pre-1.0 — medium/high:** upstream warns that compatibility can change and does not recommend PocketBase for critical production systems. Mitigate by pinning v0.40.4, keeping migrations, reviewing release notes before upgrades, and using this only as a feedback pilot.
- **Data continuity — medium:** SQLite lives in `/data`; the Dockerfile declares a volume, but backups and a restore drill are not configured. Do not replace/remove the data volume; verify it on the Dokploy host before relying on long-term feedback.
- **Account recovery — medium:** SMTP, email verification and password reset are not configured. Testers must remember their own password; do not share accounts.
- **Demo evaluation — high:** matching words is not a meaningful learning assessment. Keep `mock · heuristic-demo-v1` disclosed and avoid teaching claims based on its score.
- **Generic lesson scaffolds — medium:** routes provide structure, not subject-specific instruction. Keep this explicit until generated content and its quality controls exist.
- **Unconfigured AI provider — high:** OpenCode contract and credentials are not integrated; do not send user data to an unconfigured provider.
- **Voice premature — low:** keep voice out of scope until core feedback is validated.
