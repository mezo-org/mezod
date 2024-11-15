<!-- 
Pin the issue this PR 'References' to. 

You can use both GitHub and Linear issues. If an issue exists in both 
systems, always prefer the GitHub issue as it lives closer to the pull request.

For GitHub, use the issue number (e.g. #100) or the full issue URL 
(e.g. https://github.com/<organization>/<repository>/issues/100).

For Linear, use the issue ID (e.g. ENG-100) or the full issue URL
(e.g. https://linear.app/<workspace>/issue/ENG-123/<title>).

Use 'Closes' instead of 'References' if this PR should close the issue when 
merged.
-->
References: <!-- Put GitHub/Linear issue here -->

<!--
Add 'Depends on' if this PR depends on another PR. Remember about setting this 
PR's base branch to the branch of the PR it depends on. Keep this PR as draft
until the PR it depends on is merged.
-->

<!--
If this PR is a work in progress, mark it as a draft. In such a case, 
the minimum is filling out the 'Introduction' section. If possible, placing a 
TODO list with planned changes and current progress in the 'Changes' section
is strongly recommended.
-->

### Introduction

<!-- 
Add a short introduction describing the context and the goal of this PR.
-->

### Changes

<!--
Describe specific changes made in this PR. Use level-4 headings to separate
different sections of changes. For example:

#### The new XXX component
(...)

#### Changes in the YYY module
(...)
-->

### Testing

<!--
Describe how the presented changes can be tested. Execute some basic tests
on your own and provide a short summary of the results in this section.
-->

---

### Author's checklist

- [ ] Provided the appropriate description of the pull request
- [ ] Updated relevant unit and integration tests
- [ ] Updated relevant documentation (`docs/`) or specification (`x/<module>/spec/`)
- [ ] Assigned myself in the `Assignees` field
- [ ] Assigned `mezod-developers` in the `Reviewers` field and notified them on Discord

### Reviewer's checklist

- [ ] Confirmed all author's checklist items have been addressed
- [ ] Considered security implications of the code changes
- [ ] Considered performance implications of the code changes
- [ ] Tested the changes and summarized covered scenarios and results in a comment
