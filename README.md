# Zen 

Zen is an application for structuring your daily routine.

![zen-logo](/web/src/lib/assets/favicon.svg)

> [!NOTE]  
> Zen has no feature to read your mind yet; therefore, cheating your way up the leaderboard is possible.
> However, keep in mind that doing so does not outsmart the developers, but yourself.

## Deployment

**Tool prerequisites:**

- `pulumi`
- `go`
- `npm/nodejs`
- `awscli`


**AWS prerequisites:**

- Create a route53 public hosted zone for your domains if you want `monk` to automatically setup all required records. 
- To make the Email verification work correctly your AWS account must be [unsandboxed from SES](https://docs.aws.amazon.com/ses/latest/dg/request-production-access.html).


The deployment of zen is facilitated via `monk` 🪬, a tiny cli tool that wraps the underlying pulumi process which builds and deploys the software. The process is interactive, so the only command you ever need (for both deployment and upgrades) is this:

```bash
go run cmd/monk/monk.go
```

> [!IMPORTANT]
> If you are not me, you should change the [privacy policy](/web/src/routes/privacy-policy/+page.svelte) and [terms of service](/web/src/routes/privacy-policy/+page.svelte) before deploying. 


## Technical Information
---

Zen was built to run highly available on an extremely low-cost cloud-infrastructure while allowing *me* to ship changes very quickly.
This comes with certain drawbacks: 

- Zen is based on various eventually consistent mechanisms (e.g. timing updates from the weekend might be added to next weeks leaderboard). Certain operations might seem to be magically swallowed or disappear in hyperedgecases.
- The leaderboard cannot be queried server-side; you must load the full json and sort it client-side (+ the json is aggressively cached by cloudfront).
- The frontend planner is dogwater spaghetti code. Right now this keeps it very simple and maintainable for **me*** (due to very loosely requirements).

For this particular scope, this is totally fine; however, please do NOT take this as an example or argument of how to correctly implement business-critical software!
