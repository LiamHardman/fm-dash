import dash_html_components as html
import dash_core_components as dcc

import header_footer


def faq_page_layout():
    return html.Div(
        [
            header_footer.create_header(),
            html.Div(
                className="faq-container",
                children=[
                    html.H1("Frequently Asked Questions", className="faq-title"),
                    dcc.Markdown(
                        """
                    ----
                    ## How are scores calculated?

                    **Category Scores:** Every players' attributes get split into Physical, Mental and Technical. This then gets multiplied by numbers defined by each 'weight' , which is mostly determined by the best attributes for a given role. The number at the end of this calculation is then ready to be used for scoring!



                    **Normalization:** The scores in each category are normalized on a scale from 0 to 100. This ensures a fair comparison between players, regardless of different scoring ranges in each category. This means that whether you're in non-league or challenging for the champions league, you'll still see the full 0-100 overall range in use.



                    **Overall Score:** This score is the average of the normalized scores in the three categories. It gives a holistic view of a player's abilities.



                    **Value Score:** This score is a more complex calculation that takes into account the player's overall score and what their transfer market value is. It's a way to understand a player's worth in a financial context.



                    **Top Players Selection:** Finally, the application filters and presents the top 500 players based on their overall scores.

                    ## What's the current usage limits?
                    
                    You can analyze as often as you like, and there's no limits for each person, IP etc.
                    There's a max file size limit which will be changed as the app's performance is monitored. Currently, it's 15MB.
                    In player counts, that's roughly 12000 players. If you want more than that, you can host yourself, and change the upload_max in the config.yml file to whatever you'd like!

                    ## What do you do with my data?

                    Not much at all! At the moment, we store some information such as file sizes that are uploaded and file names (but not contents!) for a short period of time. This is just to figure out how the app's performing as load changes.
                    No PII's stored by the app, and we don't use any cookies or other tracking methods.

                    ## This is free, but is there a catch? 
                    
                    I'd like to think not. This was made as a passion project to learn more about Python development.

                    ## What if I want a new feature or have a suggestion?

                    You can either check out the discord (linked in the footer) , or create an issue on GitHub. We're always looking for ways to improve the application, so any feedback is welcome!

                """,
                        className="faq-text",
                    ),
                ],
                style={
                    "padding": "20px",
                    "background-color": "#1E293B",
                    "color": "white",
                },
            ),
            header_footer.create_footer(),
        ],
        style={
            "background-color": "#1E293B",
            "color": "#FFFFFF",
            "height": "100vh",
            "display": "flex",
            "flex-direction": "column",
            "justify-content": "space-between",
        },
    )
