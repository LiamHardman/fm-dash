# FM-Dash

[FM-Dash](https://dev.fm-dash.com) is a Football Manager Data analysis tool made with the aim of providing a number to the 'eye-test' that many FM players rely on to make signings. While you can use the above website to access this tool, you can host it yourself too!



While tools like FM Genie Scout and FMRTE use and provide data that players otherwise couldn't see, this only uses features and data within the game itself, and provides scores on a relative basis. That means this tool (should) scale well no matter what level in the football pyramid you're in.

## Feature Set

- Grade players based on a weighted attribute criteria, and display a score from 0-100 for Mental, Physical, and Technical skills to make up an overall score
- Generate a Value Score based on a player overall score and transfer value.
- Compare players with similar attributes to each other by clicking on the player in the data table.
- Visualize players in a detailled Player Info tab featuring in-game attributes, performance stats and a FIFA-style card vizualisation.


## Usage
- Firstly, you'll need to import the `fm-dash-search.fmf` scouting view. Hop onto the scouting screen, then click on General Info (or whatever your view is called) -> Import View . Navigate to the location of where you placed `fm-dash-search.fmf`. Then, you'll see lots of attributes next to each player.
  - If you want to run your squad data through this, you can use the `fm-dash-squad-view.fmf` instead.

- The FM UI is a bit finicky, but you *should* be able to click on the empty space around a player in the scouting view , hold CTRL-A to select all players and CTRL+P to save to file. **Web page** should be chosen, since this is the only compatible format with this (and other) FM analysis tools.

- Navigate over to FM-Dash, and upload this export file and select a "Weight Set". Then, you'll be presented with a graph along with a table of players.
  - The 'weight set' will prioritize certain attributes to search for, but it doesn't directly filter anyone out. For example, if you pick the "DC - Generic" weight set, you can have attackers show up in this as well, albeit with a lower score.

  - Searching in the tables is intuitive. You can search for specific positions and/or <23 for example to show under 23 players that play in certain positions.
  - The graph will always sort values descending on the X axis.

- You can click on any player and you'll open the "Find Similar" table. This'll find players that have at least 75% of their stats within 2 of the player you've selected. It'll then return a "similarity score". This is just how close they are to the player you've selected. A player with a lower similarity score may actually be a better player.

- If you click on "View Details" , you can find the players attributes and performance stats in some handy bar charts.

- You can also now click on "Use Player Attributes as Weights". This is perfect for if you have a particularly unique player with a set of stats you'd like to find in another player.
- Player and club images retrieved automatically from SoFIFA and SortitoutSI respectively. Regen faces retrieved from local disk (link [here](http://api.fm-dash.com/regen_facepack/regen.zip) to download) or from my own web mirror.




## Limitations

- Goalkeeper attributes aren't supported in the search. I do plan to add this in the future. For now though, goalkeepers will have a very low overall no matter what weight you use, and the detailled view will likely not work.
- If you have a slow internet connection, rendering player and club images with non-regens will be quite slow. This is because FM-Dash fetches data from both SortitoutSI and SoFIFA for club/league and player images respectively.
- Exports using the default view isn't supported. `fm-dash-search.fmf` must be used for full feature support. Alternative views *may* work, but they're not supported.
- Only HTML data exports are supported.
- Analyzing over ~8k players on the [FM-Dash site](https://dev.fm-dash.com) does slow things down considerably, and that's the limit as it stands. As a general use pattern, only export players that are worth at least 1% of your most valuable player, and are at least 'Doubtful' interest.
- If you're hosting this yourself in a container environment, watch out for the somewhat large image pull.
- The GitHub repo itself doesn't include the NewGAN facepack. You'll need to source that yourslf, extract the folders into `fifa_card_assets/regen` and run `fifa_card_assets/sort_regen_images.py`to organize them into the correct folder structure.




## Requirements & Installation

### Hardware System Requirements

| Component | Requirement                  | Notes                                           |
| --------- | ---------------------------- | ----------------------------------------------- |
| CPU       | x86 2Ghz Dual Core           | Untested on ARM                                 |
| RAM       | 4GB (Linux) - 6GB (Windows)  | 1000MB Free RAM for >2k players, 1500MB for 2-20k |
| OS        | Windows 7+ , Linux kernel 3+ |                                                 |
| Software  | Python 3.8+                  |                                                 |


### Running Locally

You can clone this repo and run `pip install -r requirements.txt`. When that's run, you can then launch the app by running `uwsgi --ini uwsgi.ini --http-socket=127.0.0.1:8080`(or yourip:8080 to expose it on non-local machines!). Then you can just navigate to http://127.0.0.1:8080 in your browser, and the app will now be accessible locally!


### Docker / Docker-Compose [RECOMMENDED]

The docker image for this is available on [Docker Hub](https://hub.docker.com/r/lhitn/fm-dash), and is the recommended way to install and use FM-Dash.
You can also either build this yourself by cloning this repo and running `docker build . -t lhitn/fm-dash && docker run -p 8080:80 lhitn/fm-dash` or by using docker-compose:

```yaml
version: "3.8"
services:
  fm-dash:
    image: lhitn/fm-dash:latest
    ports:
      - 8080:80
```



## Contribution, Credit & Project Guidelines

I made this project to learn more about Python along with full-stack programming as a whole, and I was initially heavily inspired by[Squirrel Plays](https://www.youtube.com/@squirrel_plays_fof4318) and how he went about things. Things took a bit of a turn, and I got hooked on making this for a couple months straight. I've no doubts people will be able to find ways to improve parts of this, and if you're questioning anything, please let me know and I'll answer and document where I can further.


### Credits:

- Zealand's team for the [NewGAN Facepack](https://fm-base.co.uk/resources/newgan-facepack.1266/)
- SoFIFA (Player Information API Search)
- SortitOutSI (League and Club API Search)


### Contribution & Issue Guidelines

As I hinted at previously, I'm more than open to contributions as this is my first project, so I'm still learning (a lot). Only strict requirements are to ensure that whether you're fixing or reporting a bug, you're as clear as possible. Expect some healthy discussion, and lots of questions from myself. Any dev-focused docs will be in `contrib-docs/` until I find a better location for them.