---
description: "Use when reviewing generated SPDX license range changes, especially additions or removals of complete license range families."
applyTo: "spdxexp/spdxlicenses/license_ranges.go"
---

# License Range Family Review

- Treat the addition or removal of a complete license range family as requiring human review.
- Always leave an inline review comment on the added or removed family block, or on the nearest changed line when the family was removed. Start the comment with **Human review required: license range family changed** and identify the family.
- Ask the human reviewer to verify that the family change is intended and that its version grouping and ordering are correct.
- Compare the diff with the **License range families** summary in the pull request description and report any discrepancy.
- Do not leave this comment for a version added to or removed from an existing family unless the grouping or ordering is incorrect.