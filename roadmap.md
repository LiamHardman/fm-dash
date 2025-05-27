
# Project Roadmap

This document outlines the planned features, improvements, and tasks for the project. It is organized by priority, reflecting the general order in which we aim to address these items.

## Near Term / High Priority

These are tasks that are crucial for the immediate development and stability of the project.

* Adjust the leagues view to display star ratings only, removing detailed attribute ratings (ATT/MID/DEF).
* Integrate backend log collection with Signoz.
* Conduct thorough testing of the Signoz logging integration.
* Expand tracing capabilities and metric collection throughout the application.
* Implement Signoz setup for client-side (JavaScript) event tracking.
* Ensure trace IDs are consistently included in all relevant logs.
* Investigate and resolve errors occurring when users interact with metrics.
* Expand and improve the "Getting Started" guide and overall documentation.

## Medium Term / Medium Priority

These items are important for enhancing the project's functionality and user experience.

* Design or select a new logo for the project.
* Address issues with the PowerShell development script to ensure it functions as intended; potentially create separate start and stop scripts.
* Overhaul the user interface for the dataset display page.
* Redesign the team page UI, with an emphasis on star ratings, player ratings, and displaying the team's league.
* Recalibrate the overall player rating system to prevent inflation for mid-tier players.
* Explore and prototype a player display format inspired by FIFA player cards.
* Research methods to extract player division and competition data from HTML exports.
* Implement support for Football Manager saves where player attributes are masked.

## Future / Low Priority

These are ideas and tasks that will be considered for future development cycles.

* Introduce a feature to configure data retention policies, including an option to disable automatic deletion.
* Refresh the demo dataset to align with the latest application views and features.
* Brainstorm and decide on a new, more distinct name for the project.
* Create detailed documentation explaining the methodologies behind specific calculations, such as the FIFA-style statistics and overall player ratings.
* Incorporate performance percentile calculations specifically for goalkeepers.
* Explore the feasibility of adding player images, contingent on the availability of unique player IDs in exports.
* Undertake a comprehensive review and potential redesign of the overall user interface.

---

*This roadmap is a living document and will be updated as the project evolves.*
