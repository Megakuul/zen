# Zen 

Zen is an application for structuring your daily routine.

![zen-logo](/web/src/lib/assets/favicon.svg)

> [!NOTE]  
> Zen has no feature to read your mind yet; therefore, cheating your way up the leaderboard is possible.
> However, keep in mind that doing so does not outsmart the developers, but yourself.

## Technical Information
---

Zen was built to solve a specific problem while running on high available but extremly low-cost cloud-infrastructure.
This comes with certain drawbacks: 

- Zen is based on various eventually consistent mechanisms (e.g. timing updates from the weekend might be added to next weeks leaderboard). Certain operations might seem to be magically swallowed or disappear in hyperedgecases.
- The leaderboard cannot be queried server-side; you must load the full json and sort it client-side (+ the json is aggressively cached by cloudfront).

For this particular scope, this is totally fine; however, please do NOT take this as an example or argument of how to correctly implement business-critical software!
