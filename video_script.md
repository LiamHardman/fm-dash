Hey everyone, hope you're all doing well. My name's Liam, and I'd like to welcome you to FM Dash. It's been my passion project for a little while, and while there's a few bits I want to improve, it's now at the stage where I'm happy to show off what it can do. Hopefully you can have as much fun using it as I have had making it so far. I was inspired by Squirrel Plays and his analysis of FM players in Python, and up until then, I'd no idea you could perform data exports in Football Manager. When this goes out for full release in January, it'll be compatible with any desktop operating system as well as Docker, but you can start using the publicly available version linked in the description. For now, the max file size on the public version is 10MB or about 8000 players, but I'll be reviewing this over time and of course you can do what you want with the local version when it's ready. There's still a bit of cleanup needed, but this is going to be available on GitHub as soon as possible.

Just to reiterate, this is completely free, there's no signup needed and while you can download the views on the website, you can also use the workshop links in the description.

I'll be splitting this into 2 different sections: Using the app, and setting up the Football Manager views and export. With that said, let's get into the app. Click on Get Started, and Upload Data. I've got an FM data export already, so let's upload it. I've not actually looked at this yet, so we can try some live scouting.


 There's a decent bit to take in, but let's start off with the graph. You can change the number of results to show, X and Y axis and the graph type. I'd recommend leaving the graph type as-is, and while you can see performance stats such as xG and Tackles per 90, I've got another view that's great for that which I'll show you later on.
This is mainly used to find the statistical outliers that will be harder to spot when scrolling through a table of data. There's a lot of plots on this graph, but you can hover over any player and see some handy information.

You can also sort by value score to see who isn't necessarily the best player, but if they're good enough for your team, they'll be a great buy.

Next up, let's take a look at the table and I'll address the biggest elephant in the room. We've got some information that's in the data export by default, but I've also got Overall, Physical, Mental, Technical and Value Scores. This is where FM Dash diverts from tools like FMRTE and FM Scout Editor quite a lot. These scores aren't directly based on Current Ability, and they're not really comparable. The main reason for this is that the overall is calculated using a set of weighted attributes which you can see at any point by clicking on Weights Info.


All of these scores are also relative, and that's what gives it quite a unique spin. If I was a non-league side and exported players that were interested in joining me, I'd still see players with overalls in the high 80's or 90's. The weights themselves are still very much a work in progress since there's not too much information out there about the exact impact certain attributes have for every role, but there's enough about some roles that I've been able to approximate and get decent results. If you end up finding some oddities while testing yourself though, please let me know in the discord channel linked in the description and the footer of the site.



 No matter how many players you import , you'll see a max of 500 at any one time, so taking advantage of the filters can be quite important.
The one that you'll likely be using the most is the Role Weighting. This doesn't filter out anyone, but it does change the way that the scores are calculated. If I select, say , Wide Centre Back, it's going to show players that can play elsewhere, but if you'd like, you can also show only certain positions.

All of the filters you see can be used in conjunction with each other, and I'd recommend you play around with these as much as possible. If this seems a bit intimidating, I'll show you a handy alternative.

Naturally, a lot of people use tools like this to simply find a better player for their team, or a replacement for someone leaning for a move elsewhere. That's why you can use the Upgrade Finder. The only things you have to put in here are club, position and role. The rest is up to you.  By default, this only shows players that are at least 3 overall better than your best player, but you can change this if you'd like. For now, let's use Newcastle as the example club and let's try to find a new top centre-mid.

As you can see, we've got a few options. This is sorted by value score by default, but feel free to mess with this. Ideally, you should set the filters up to get some results show, but less than a page so then you see the absolute ideal signings.

What I found while using this app a while ago was that there was often a middle ground, You have a unique player that you need to replace so you need to go over things with a more fine tooth comb, but you also don't want to spend loads of time finding that perfect replacement. This is where using a players attributes as weights comes in handy. (pick someone thats relatively unique) So this is more useful for when you have a player that really excels in a certain area. What I've personally found is that if you combine this with sorting by the youngest players as well, you can find some real gems.

You'll notice when you press on any player, you'll see another table pop up. This has varying levels of usefulness depending on the context, but this'll show up the most similar players noted by a similarity score. This pretty much only matches for players with mostly similar attributes, so this can be handy for a really quick view of who might be an alternative target than who you're currently looking at.

Now all of this is pretty cool, but there's not much point if you have to keep switching between Football Manager and FM Dash. That's where the player info tab comes in handy. Click on any player and then View Player Info.

As you can see, we've essentially got the player attributes laid out, but we've also got their performance stats. I got inspired by the layout FBRef uses and it's a really good way to see who's ready for the next step up or really outperforming everyone else in your imoprt. You can change the set of positions to compare against as well so you can get more of an understanding of where they compare with their peers rather than people in every position.

Now you'll also see an interesting visualisation at the top right. You'd be surprised how tricky this was to pull off, and it's still not perfect, but it does at least satisfy the FIFA crave that some of us have without having to actually play it. By default, player faces and badges aren't rendered, but you can simply enable these at the bottom. These are fetched from SOFIFA's API and SortItOutSI respectively and then cached so after enough time these should load really quickly.

If you want to see your whole team rendered out in these FIFA-style cards, you can simply select your formation and if you want to render a bench, and then click Generate Cards for Club. You can also change the card type and the number of times that player's upgraded. All of the upgrades are relatively randomized but still should mostly make sense. The card type also gets considered when generating the cards for your club. If you want to render your team out as all icons, feel free!

So, that's all for features, let's get into preparing FM for exporting data.

So if you haven't already, download the scouting and or squad views linked in the description below and place them into the folders onscreen depending on your operating system (Edit this in) .

Then, for the squad view you can go to Squad -> Import View -> fm-dash-squad-view

For the scouting view, you can go to scouting -> import view -> fm-dash-search. I'd also recommend you untick exclude club players here.

Now to export your data, it's somewhat finicky the first time you do it, but it's easy to get used to. If you've just imported the view, you might need to go to your home screen and back, but then just click on a player row and hold CTRL+A on Windows or COMMAND+A on Mac and then CTRL+P or Command+P . You'll see 3 options to save the export as. You want to click web page, and save this somewhere you'll remember. Now you're ready, just head over to the link in the description and you can get started using FM-Dash.


If you've made it this far, thank you very much for listening, and I really hope you enjoy using FM-Dash. If you're happy to give some feedback, feel free to comment here, or join the discord linked in the description. Thank you for your time. Bye for now.

**LINKS**
(link the forum post and images squirrel found in description)
Link workshop pages for views
Discord