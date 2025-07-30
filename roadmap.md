## Short Term (Days-Weeks)
- Interactive tutorial for first-time users
- Dataset table refactor, with a focus on customizability & maintainability
- Both user and maintainer-facing documentation improvements
    - Both removing old/non-relevant documentation, but also making it much easier to find documentation
- New demo dataset
- New total stat score
    - Implementation in the dataset table to understand how well-rounded players are
- Default docker compose that's well documented & well-supported along with self-hosting docs for other ways of hosting

## Medium Term (Weeks-Months)
- New 'Money ball rating'
- Authentication
    - This will enable:
        - Association of multiple datasets
        - Link datasets together
        - 1 wishlist for multiple datasets
- Team comparisons
- New 'simple' tab in the player details view
- 'Pros and cons' summary for both teams and players
- 'Playstyle' display for players
- Look into feasibility of AI player descriptions
- 'Create a team' page




## Long Term ('At Some Point')
- Regen faces
- Multi-language support
- User-facing docs on how different ratings are calculated



## Technical Roadmap
- Break down playerdatatable into multiple components
- Fix camelcase and snakecase inconsistencies
- Fix new pipeline code check fails
- Ability to define team in upgrade finder with second dataset
- Test notifications and resolve UI issues (or remove the feature)
- Increase test coverage