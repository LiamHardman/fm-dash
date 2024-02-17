# FM-Dash Quick-Start Guide

## Grabbing your FM Player Data

Grabbing your FM player data's really easy in the newer versions of football manager. All you need is a custom scouting or squad view, which I already have available for download at the top right of the screen. When you've got that, hop onto FM and follow the rest of this guide!

Firstly, you'll need to import the `fm-dash-search.fmf` scouting view. Hop onto the scouting screen, then click on General Info (or whatever your view is called) -> Import View . Navigate to the location of where you placed `fm-dash-search.fmf`. Then, you'll see lots of attributes next to each player.

The FM UI is a bit finicky, but you *should* be able to click on the empty space around a player's name in the scouting view , hold CTRL-A to select all players and CTRL+P to save to file. **Web page** should be chosen, since this is the only compatible format with this (and other) FM analysis tools.


## Using FM-Dash

Navigate over to FM-Dash, and upload this export file and select a "Weight Set". Then, you'll be presented with a graph along with a table of players. The 'weight set' will prioritize certain attributes to search for, but it doesn't directly filter anyone out. For example, if you pick the "DC - Generic" weight set, you can have attackers show up in this as well, albeit with a lower score.

Speaking of scores, it's important to note that all of the scores you'll see are relative. That means if you're in non-league and viewing your squad, you'll see people with high 90's in overall. If you had those same players in your scouting list as Real Madrid though, they likely wouldn't even show up in the top 1000 that's automatically picked out to display. What this does offer though is a way to see who's going to improve your squad no matter where you are in the footballing pyramid. Just note that you can't really compare scores from two different datasets to each other.

Searching in the tables is quite intutive. You can search for specific positions and/or <23 for example to show under 23 players that play in certain positions. The graph will always sort values descending on the X axis. For now, you can play with any graph type and any stat. That means you can get some questionable results. Stick to non-performance stats and the scatter plot graph for the most part.

You can click on any player and you'll open the "Find Similar" table. This'll find players that have at least 75% of their stats within 2 of the player you've selected. It'll then return a "similarity score". This is just how close they are to the player you've selected. A player with a lower similarity score may actually be a better player.

If you click on "View Details" , you can find the players attributes and performance stats in some handy bar charts.

You can also now click on "Use Player Attributes as Weights". This is perfect for if you have a particularly unique player with a set of stats you'd like to find in another player.

Now let's say you want to ignore everything you've read above, and just want to find someone that'll improve your team. Thankfully, you can do just that. Click on "Upgrade Finder" , input your club, position you want to find an upgrade for, and the role they play, and let us do the maths. You can also optionally set a max age and minimum overall upgrade required to show in your results.


## Something's broken!

Tell me about it please! You can click on the discord button below to join the FM-Dash server. Stick something in the #bug-reports channel and I'll get it fixed.


## I need some help!

That's what the #help channel is for! Fire over a quick message and we can sort out what you need help with.