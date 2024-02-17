# Scoring System Summary

## Weights

This is what separates SI's 'Current Ability' calculation from how FM-Dash does things. While FM will have their own weights system for each role in the background, this isn't shown transparently to the user. Meanwhile FM-Dash lets you choose your own weights, and you can view how stats are weighted at /weights. You can also click on "Use Player Attributes as Weights" and it'll use the stats of the player you've got selected as the 'weight'. This is particularly handy if you're not necessarily trying to find someone *similar* to that player, but someone who is better at the same role that the player you've selected is currently playing in your team.


## Overall, Physical and Mental Scores

Each attribute for a given player is multiplied by the attribute associated with the 'weight'. For example, if you searched using the DC - BPD weight and viewed William Saliba, you'd see that his 14 acceleration would be multiplied by 90. All of those multiplied attributes in a given category are then divided by the number of them (8 for Physical, 14 for Mental & Technical). That'll give you the scores for each category. The overall is just them 3 put together, and divided by 3 again. We're not quite done there technically, but that's enough to know for now.


## Value Scores

Naturally, better players are more expensive. At the elite-level, the increase can get close to exponential. Thankfully, FM-Dash's 'Value Score' aims to help people find good buys in any league. Essentially, this'll check the price of a player along with its overall score and spit out an easy to read 'value score'. It only factors in Overall and Transfer Fee (For now), so you might want to not exclusively pay attention to this.

## Normalization

This is where it gets slightly more technical. It's quite difficult to make an overall & value scoring system that is both scalable and accurate. However, hopefully this is what you'll see exhibited in FM-Dash. You'll see that scores are from 0-100 (or mostly 0-60 for value score). That's no coincidence. FM-Dash will , in a given dataset, look at the highest and lowest overall scores, and ensure that the highest single score should be 100, and then it will trickle down from there. If a player is 100 rated, then someone who is 80 rated will be 80% as good when comparing with the same weight set.

One downside of this is that no two datasets are equal. However, that downside, in my opinion, is outweighed significantly by the fact that you can datasets from any level of football and get the full 0-100 rating system. This means that you can easily, and accurately, see a difference in ability between players.
