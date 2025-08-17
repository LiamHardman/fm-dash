// Playstyle taglines and attribute mapping for FM-style attributes
// Provides a mapping of position short codes to archetypal playstyles along with
// their significance and the FM attribute names that best represent the style.

export const ATTRIBUTE_NAME_TO_KEY = {
  // Technical
  Corners: 'Cor',
  Crossing: 'Cro',
  Dribbling: 'Dri',
  Finishing: 'Fin',
  'First Touch': 'Fir',
  'Free Kicks': 'Fre',
  'Free Kick Taking': 'Fre',
  Heading: 'Hea',
  'Long Shots': 'Lon',
  'Long Throws': 'L Th',
  Marking: 'Mar',
  Passing: 'Pas',
  'Penalty Taking': 'Pen',
  Tackling: 'Tck',
  Technique: 'Tec',

  // Mental
  Aggression: 'Agg',
  Anticipation: 'Ant',
  Bravery: 'Bra',
  Composure: 'Cmp',
  Concentration: 'Cnt',
  Decisions: 'Dec',
  Determination: 'Det',
  Flair: 'Fla',
  Leadership: 'Ldr',
  'Off the Ball': 'OtB',
  Positioning: 'Pos',
  Teamwork: 'Tea',
  Vision: 'Vis',
  'Work Rate': 'Wor',

  // Physical
  Acceleration: 'Acc',
  Agility: 'Agi',
  Balance: 'Bal',
  'Jumping Reach': 'Jum',
  'Natural Fitness': 'Nat',
  Pace: 'Pac',
  Stamina: 'Sta',
  Strength: 'Str',

  // Goalkeeping
  'Aerial Reach': 'Aer',
  'Command of Area': 'Cmd',
  Communication: 'Com',
  Eccentricity: 'Ecc',
  Handling: 'Han',
  Kicking: 'Kic',
  'One on Ones': '1v1',
  'Punching (Tendency)': 'Pun',
  Reflexes: 'Ref',
  'Rushing Out (Tendency)': 'TRO',
  Throwing: 'Thr',
}

// Note: Some provided attribute names may not exist in the dataset (e.g., "Recovery").
// Those will be ignored when scoring.

export const PLAYSTYLE_TAGLINES = {
  GK: [
    {
      playstyle: 'Sweeper-Keeper',
      significance:
        'Acts as an eleventh outfield player, comfortable coming off their line to intercept through balls and initiate attacks from the back.',
      fm_attributes: ['Rushing Out (Tendency)', 'Passing', 'First Touch', 'Vision', 'Anticipation'],
    },
    {
      playstyle: 'Shot-Stopper',
      significance:
        'A traditional goalkeeper whose primary strength is making exceptional saves, often through pure reflexes and agility.',
      fm_attributes: ['Reflexes', 'One on Ones', 'Handling', 'Agility', 'Positioning'],
    },
    {
      playstyle: 'Commanding Presence',
      significance:
        'Dominates their penalty area, confidently claiming crosses and organizing the defense. A vocal and physically imposing figure.',
      fm_attributes: ['Command of Area', 'Aerial Reach', 'Communication', 'Handling', 'Strength'],
    },
    {
      playstyle: 'Acrobatic Goalie',
      significance:
        'Known for making spectacular, unorthodox, and visually impressive saves. Relies heavily on agility and flexibility.',
      fm_attributes: ['Agility', 'Reflexes', 'Flair'],
    },
    {
      playstyle: 'Distribution Specialist',
      significance:
        'Excels at distribution, capable of launching precise attacks with long kicks or accurate throws.',
      fm_attributes: ['Kicking', 'Passing', 'Throwing', 'Vision', 'Technique'],
    },
    {
      playstyle: 'Line-Keeper',
      significance:
        'A more conservative keeper who rarely strays from the goal line, relying on excellent positioning and reflexes to make saves.',
      fm_attributes: ['Positioning', 'Reflexes', 'Handling', 'Concentration'],
    },
    {
      playstyle: 'Penalty Specialist',
      significance:
        'Has a knack for saving penalties, often through a combination of psychological games, anticipation, and explosive reflexes.',
      fm_attributes: ['Reflexes', 'Composure', 'Anticipation', 'Concentration'],
    },
  ],
  SW: [
    {
      playstyle: 'Classic Sweeper',
      significance:
        'A purely defensive player who reads the game from behind the defensive line, intercepting through balls and covering for other defenders.',
      fm_attributes: ['Anticipation', 'Positioning', 'Concentration', 'Decisions', 'Pace'],
    },
    {
      playstyle: 'Libero',
      significance:
        'A modern, attacking sweeper who steps out from the backline with the ball, acting as an extra playmaker and joining the attack.',
      fm_attributes: ['Dribbling', 'Passing', 'Vision', 'Anticipation', 'Flair'],
    },
    {
      playstyle: 'Ball-Playing Sweeper',
      significance:
        "Operates as the starting point of attacks from deep, using exceptional passing range to bypass the opposition's press without needing to dribble forward.",
      fm_attributes: ['Passing', 'Vision', 'Technique', 'First Touch', 'Composure'],
    },
    {
      playstyle: 'Aggressive Sweeper',
      significance:
        'Proactively steps up to challenge attackers and make interceptions, relying on timing and aggression to win the ball.',
      fm_attributes: ['Aggression', 'Tackling', 'Pace', 'Anticipation'],
    },
    {
      playstyle: 'Enforcer',
      significance:
        'A physically imposing sweeper who prioritizes intimidating opponents and winning physical battles above all else.',
      fm_attributes: ['Strength', 'Aggression', 'Bravery', 'Tackling'],
    },
    {
      playstyle: 'Pacy Sweeper',
      significance:
        'Employs exceptional speed to cover for a high defensive line, capable of winning races against fast strikers.',
      fm_attributes: ['Pace', 'Acceleration', 'Anticipation', 'Recovery'],
    },
    {
      playstyle: 'Tactical Sweeper',
      significance:
        'Acts as the brain of the defense, organizing the backline, communicating constantly, and ensuring tactical discipline.',
      fm_attributes: ['Communication', 'Leadership', 'Decisions', 'Positioning', 'Teamwork'],
    },
  ],
  DC: [
    {
      playstyle: 'Build-Up Defender',
      significance:
        'A modern defender who is elegant and comfortable in possession, capable of breaking lines with passes or carrying the ball into midfield.',
      fm_attributes: ['Passing', 'First Touch', 'Vision', 'Composure', 'Technique'],
    },
    {
      playstyle: 'Traditional Stopper',
      significance:
        'A rugged, old-school defender who prioritizes defensive solidity above all else. Focuses on tackles, clearances, and simple play.',
      fm_attributes: ['Tackling', 'Marking', 'Positioning', 'Bravery'],
    },
    {
      playstyle: 'Bruiser',
      significance:
        'A physically dominant defender who excels at overpowering forwards, winning headers, and making strong challenges.',
      fm_attributes: ['Strength', 'Aggression', 'Jumping Reach', 'Bravery'],
    },
    {
      playstyle: 'Covering Defender',
      significance:
        'Reads the game intelligently to sweep up behind the defensive line, using pace and anticipation to intercept through balls.',
      fm_attributes: ['Pace', 'Acceleration', 'Anticipation', 'Positioning', 'Concentration'],
    },
    {
      playstyle: 'Wide Centre-Back',
      significance:
        'In a back three, this defender overlaps into wide areas, providing an extra attacking threat similar to a full-back.',
      fm_attributes: ['Dribbling', 'Crossing', 'Work Rate', 'Pace'],
    },
    {
      playstyle: 'Aerial Dominator',
      significance:
        'Commands the air in both penalty boxes, winning defensive headers and posing a significant threat from offensive set-pieces.',
      fm_attributes: ['Jumping Reach', 'Heading', 'Strength', 'Balance'],
    },
    {
      playstyle: 'Intelligent Defender',
      significance:
        'Relies on an expert reading of the game rather than pure physicality, excelling at interceptions and perfect positioning.',
      fm_attributes: ['Anticipation', 'Positioning', 'Concentration', 'Decisions', 'Composure'],
    },
  ],
  DR: [
    {
      playstyle: 'Attacking Full-back',
      significance:
        'Provides width high up the pitch, constantly overlapping the winger to deliver crosses and join the attack.',
      fm_attributes: ['Crossing', 'Dribbling', 'Work Rate', 'Stamina', 'Pace'],
    },
    {
      playstyle: 'Defensive Full-back',
      significance:
        'A cautious full-back whose primary responsibility is to nullify the opposing winger and protect the flank.',
      fm_attributes: ['Tackling', 'Marking', 'Positioning', 'Concentration'],
    },
    {
      playstyle: 'Inverted Full-back',
      significance:
        'Drifts into central midfield when the team has possession, creating overloads and acting as an auxiliary playmaker.',
      fm_attributes: ['Passing', 'Vision', 'First Touch', 'Decisions', 'Teamwork'],
    },
    {
      playstyle: 'Complete Full-back',
      significance:
        'An all-action player who is a major contributor to both defense and attack, requiring exceptional physical and technical attributes.',
      fm_attributes: ['Work Rate', 'Stamina', 'Crossing', 'Tackling', 'Pace'],
    },
    {
      playstyle: 'Pacy Overlapper',
      significance:
        'Uses blistering speed as their main weapon, burning past opponents to get into dangerous positions on the counter-attack or in build-up play.',
      fm_attributes: ['Pace', 'Acceleration', 'Stamina', 'Off the Ball'],
    },
    {
      playstyle: 'Technical Full-back',
      significance:
        'Comfortable in tight spaces and under pressure, able to play out from the back and contribute to a possession-based style.',
      fm_attributes: ['First Touch', 'Dribbling', 'Passing', 'Technique', 'Composure'],
    },
    {
      playstyle: 'Tackling Machine',
      significance:
        'A defensive specialist who excels at winning the ball back, focusing almost entirely on their defensive duties.',
      fm_attributes: ['Tackling', 'Positioning', 'Aggression', 'Anticipation'],
    },
  ],
  DL: [
    {
      playstyle: 'Attacking Full-back',
      significance:
        'Provides width high up the pitch, constantly overlapping the winger to deliver crosses and join the attack.',
      fm_attributes: ['Crossing', 'Dribbling', 'Work Rate', 'Stamina', 'Pace'],
    },
    {
      playstyle: 'Defensive Full-back',
      significance:
        'A cautious full-back whose primary responsibility is to nullify the opposing winger and protect the flank.',
      fm_attributes: ['Tackling', 'Marking', 'Positioning', 'Concentration'],
    },
    {
      playstyle: 'Inverted Full-back',
      significance:
        'Drifts into central midfield when the team has possession, creating overloads and acting as an auxiliary playmaker.',
      fm_attributes: ['Passing', 'Vision', 'First Touch', 'Decisions', 'Teamwork'],
    },
    {
      playstyle: 'Complete Full-back',
      significance:
        'An all-action player who is a major contributor to both defense and attack, requiring exceptional physical and technical attributes.',
      fm_attributes: ['Work Rate', 'Stamina', 'Crossing', 'Tackling', 'Pace'],
    },
    {
      playstyle: 'Pacy Overlapper',
      significance:
        'Uses blistering speed as their main weapon, burning past opponents to get into dangerous positions on the counter-attack or in build-up play.',
      fm_attributes: ['Pace', 'Acceleration', 'Stamina', 'Off the Ball'],
    },
    {
      playstyle: 'Technical Full-back',
      significance:
        'Comfortable in tight spaces and under pressure, able to play out from the back and contribute to a possession-based style.',
      fm_attributes: ['First Touch', 'Dribbling', 'Passing', 'Technique', 'Composure'],
    },
    {
      playstyle: 'Tackling Machine',
      significance:
        'A defensive specialist who excels at winning the ball back, focusing almost entirely on their defensive duties.',
      fm_attributes: ['Tackling', 'Positioning', 'Aggression', 'Anticipation'],
    },
  ],
  WBR: [
    {
      playstyle: 'Complete Wing-Back',
      significance:
        'An all-action player who is a major contributor to both defense and attack, requiring exceptional physical and technical attributes.',
      fm_attributes: ['Work Rate', 'Stamina', 'Crossing', 'Tackling', 'Pace'],
    },
    {
      playstyle: 'Attacking Wing-Back',
      significance:
        "Functions more like a winger than a defender, tasked with providing the team's primary attacking width and delivering dangerous crosses.",
      fm_attributes: ['Crossing', 'Dribbling', 'Pace', 'Off the Ball', 'Flair'],
    },
    {
      playstyle: 'Inverted Wing-Back',
      significance:
        'Cuts into central midfield from a wide position when the team is in possession, acting as an extra midfielder to overload the opposition.',
      fm_attributes: ['Passing', 'Vision', 'Decisions', 'Teamwork', 'Anticipation'],
    },
    {
      playstyle: 'Defensive Wing-Back',
      significance:
        'A more conservative wing-back who prioritizes defensive duties but is still expected to get forward and support the attack.',
      fm_attributes: ['Positioning', 'Tackling', 'Work Rate', 'Stamina', 'Marking'],
    },
    {
      playstyle: 'Crossing Specialist',
      significance:
        'A player whose main purpose is to deliver high-quality crosses into the box from deep or advanced wide areas.',
      fm_attributes: ['Crossing', 'Technique', 'Vision', 'Passing'],
    },
    {
      playstyle: 'Overlapping Wing-Back',
      significance:
        'Constantly looks to run beyond the winger or inside forward to create 2-vs-1 situations and get to the byline.',
      fm_attributes: ['Off the Ball', 'Pace', 'Acceleration', 'Stamina', 'Work Rate'],
    },
    {
      playstyle: 'Shuttling Wing-Back',
      significance:
        'A tireless runner who covers the entire flank, focusing on high energy and simple, effective play rather than technical flair.',
      fm_attributes: ['Work Rate', 'Stamina', 'Teamwork', 'Acceleration'],
    },
  ],
  WBL: [
    {
      playstyle: 'Complete Wing-Back',
      significance:
        'An all-action player who is a major contributor to both defense and attack, requiring exceptional physical and technical attributes.',
      fm_attributes: ['Work Rate', 'Stamina', 'Crossing', 'Tackling', 'Pace'],
    },
    {
      playstyle: 'Attacking Wing-Back',
      significance:
        "Functions more like a winger than a defender, tasked with providing the team's primary attacking width and delivering dangerous crosses.",
      fm_attributes: ['Crossing', 'Dribbling', 'Pace', 'Off the Ball', 'Flair'],
    },
    {
      playstyle: 'Inverted Wing-Back',
      significance:
        'Cuts into central midfield from a wide position when the team is in possession, acting as an extra midfielder to overload the opposition.',
      fm_attributes: ['Passing', 'Vision', 'Decisions', 'Teamwork', 'Anticipation'],
    },
    {
      playstyle: 'Defensive Wing-Back',
      significance:
        'A more conservative wing-back who prioritizes defensive duties but is still expected to get forward and support the attack.',
      fm_attributes: ['Positioning', 'Tackling', 'Work Rate', 'Stamina', 'Marking'],
    },
    {
      playstyle: 'Crossing Specialist',
      significance:
        'A player whose main purpose is to deliver high-quality crosses into the box from deep or advanced wide areas.',
      fm_attributes: ['Crossing', 'Technique', 'Vision', 'Passing'],
    },
    {
      playstyle: 'Overlapping Wing-Back',
      significance:
        'Constantly looks to run beyond the winger or inside forward to create 2-vs-1 situations and get to the byline.',
      fm_attributes: ['Off the Ball', 'Pace', 'Acceleration', 'Stamina', 'Work Rate'],
    },
    {
      playstyle: 'Shuttling Wing-Back',
      significance:
        'A tireless runner who covers the entire flank, focusing on high energy and simple, effective play rather than technical flair.',
      fm_attributes: ['Work Rate', 'Stamina', 'Teamwork', 'Acceleration'],
    },
  ],
  DM: [
    {
      playstyle: 'Anchor Man',
      significance:
        'Sits in front of the defense, shielding the backline by breaking up play and making simple passes. Purely a defensive screen.',
      fm_attributes: ['Tackling', 'Positioning', 'Marking', 'Strength', 'Concentration'],
    },
    {
      playstyle: 'Deep Orchestrator',
      significance:
        "Orchestrates the team's attacks from a deep position, dictating the tempo with a wide range of passing.",
      fm_attributes: ['Passing', 'Vision', 'First Touch', 'Composure', 'Decisions'],
    },
    {
      playstyle: 'Midfield Destroyer',
      significance:
        'An energetic and aggressive player who relentlessly hunts down the opposition to win back possession all over the midfield.',
      fm_attributes: ['Tackling', 'Work Rate', 'Aggression', 'Stamina', 'Bravery'],
    },
    {
      playstyle: 'Regista',
      significance:
        'A more aggressive version of the Deep-Lying Playmaker who moves forward with the ball, looking to create and score from deep.',
      fm_attributes: ['Passing', 'Vision', 'Dribbling', 'Flair', 'Long Shots'],
    },
    {
      playstyle: 'Half-Back',
      significance:
        'Drops between the centre-backs during build-up play, allowing the full-backs to push high and wide. A hybrid defender/midfielder.',
      fm_attributes: ['Positioning', 'Tackling', 'Passing', 'Anticipation', 'Composure'],
    },
    {
      playstyle: 'Disruptor',
      significance:
        'A combative player whose main role is to break up play and frustrate opponents through relentless pressing and aggression.',
      fm_attributes: ['Aggression', 'Work Rate', 'Bravery', 'Stamina', 'Strength'],
    },
    {
      playstyle: 'Segundo Volante',
      significance:
        "From a deep position, makes late, surging runs into the opposition's penalty area to act as a goal threat.",
      fm_attributes: ['Long Shots', 'Finishing', 'Work Rate', 'Tackling', 'Off the Ball'],
    },
  ],
  MC: [
    {
      playstyle: 'All-Action Midfielder',
      significance:
        'An all-action player with an incredible engine, contributing significantly at both ends of the pitch with tackles, runs, and goals.',
      fm_attributes: ['Work Rate', 'Stamina', 'Tackling', 'Finishing', 'Long Shots'],
    },
    {
      playstyle: 'Creative Midfielder',
      significance:
        "Moves into the space between the opposition's midfield and defense, aiming to create clear-cut chances for the strikers.",
      fm_attributes: ['Passing', 'Vision', 'First Touch', 'Dribbling', 'Technique'],
    },
    {
      playstyle: 'Mezzala',
      significance:
        'Operates in the half-spaces, making dynamic forward runs from a central position to either attack the box or drift wide.',
      fm_attributes: ['Off the Ball', 'Dribbling', 'Finishing', 'Work Rate', 'Pace'],
    },
    {
      playstyle: 'Roaming Playmaker',
      significance:
        'Given the freedom to drift around the pitch to find space, this player is the main creative force, linking play from anywhere.',
      fm_attributes: ['Off the Ball', 'Vision', 'Passing', 'Dribbling', 'Flair'],
    },
    {
      playstyle: 'Central Midfielder (Support)',
      significance:
        'A balanced midfielder who connects the defense and attack, focusing on maintaining possession and supporting teammates.',
      fm_attributes: ['Passing', 'Teamwork', 'Work Rate', 'First Touch', 'Decisions'],
    },
    {
      playstyle: 'Long-Shot Specialist',
      significance:
        'A midfielder who poses a constant threat from distance, possessing the technique and power to score spectacular goals.',
      fm_attributes: ['Long Shots', 'Finishing', 'Technique', 'Composure'],
    },
    {
      playstyle: 'Carrilero (Shuttler)',
      significance:
        'Covers the lateral ground between the two penalty boxes, linking the defensive and attacking units with tireless running and simple passes.',
      fm_attributes: ['Work Rate', 'Stamina', 'Passing', 'Tackling', 'Teamwork'],
    },
  ],
  MR: [
    {
      playstyle: 'Traditional Winger',
      significance:
        'Hugs the touchline, using pace and dribbling to beat the opposing full-back before delivering crosses into the box.',
      fm_attributes: ['Crossing', 'Dribbling', 'Pace', 'Acceleration', 'Flair'],
    },
    {
      playstyle: 'Defensive Winger',
      significance:
        'A hard-working winger whose primary duty is to support their full-back defensively, pressing the opposition and tracking back.',
      fm_attributes: ['Work Rate', 'Tackling', 'Stamina', 'Teamwork', 'Positioning'],
    },
    {
      playstyle: 'Wide Playmaker',
      significance:
        'Orchestrates play from a wide position, using exceptional vision and passing range to create chances rather than relying on pace.',
      fm_attributes: ['Vision', 'Passing', 'Crossing', 'Technique', 'Decisions'],
    },
    {
      playstyle: 'Balanced Winger',
      significance:
        "A balanced role, maintains the team's shape, links play between the full-back and forwards, and supports both defense and attack.",
      fm_attributes: ['Teamwork', 'Work Rate', 'Passing', 'Positioning', 'First Touch'],
    },
    {
      playstyle: 'Hard-Working Winger',
      significance:
        'A relentless runner who uses energy and determination to disrupt the opposition, press high, and support the team.',
      fm_attributes: ['Work Rate', 'Stamina', 'Aggression', 'Bravery'],
    },
    {
      playstyle: 'Pace Merchant',
      significance:
        'A player whose entire game is built around blistering speed, looking to outrun defenders at every opportunity.',
      fm_attributes: ['Pace', 'Acceleration', 'Off the Ball', 'Stamina'],
    },
    {
      playstyle: 'Tireless Shuttler',
      significance:
        'Runs the channel constantly, providing an outlet pass and stretching the opposition defense through sheer volume of running.',
      fm_attributes: ['Work Rate', 'Stamina', 'Off the Ball', 'Teamwork'],
    },
  ],
  ML: [
    {
      playstyle: 'Traditional Winger',
      significance:
        'Hugs the touchline, using pace and dribbling to beat the opposing full-back before delivering crosses into the box.',
      fm_attributes: ['Crossing', 'Dribbling', 'Pace', 'Acceleration', 'Flair'],
    },
    {
      playstyle: 'Defensive Winger',
      significance:
        'A hard-working winger whose primary duty is to support their full-back defensively, pressing the opposition and tracking back.',
      fm_attributes: ['Work Rate', 'Tackling', 'Stamina', 'Teamwork', 'Positioning'],
    },
    {
      playstyle: 'Wide Playmaker',
      significance:
        'Orchestrates play from a wide position, using exceptional vision and passing range to create chances rather than relying on pace.',
      fm_attributes: ['Vision', 'Passing', 'Crossing', 'Technique', 'Decisions'],
    },
    {
      playstyle: 'Balanced Winger',
      significance:
        "A balanced role, maintains the team's shape, links play between the full-back and forwards, and supports both defense and attack.",
      fm_attributes: ['Teamwork', 'Work Rate', 'Passing', 'Positioning', 'First Touch'],
    },
    {
      playstyle: 'Hard-Working Winger',
      significance:
        'A relentless runner who uses energy and determination to disrupt the opposition, press high, and support the team.',
      fm_attributes: ['Work Rate', 'Stamina', 'Aggression', 'Bravery'],
    },
    {
      playstyle: 'Pace Merchant',
      significance:
        'A player whose entire game is built around blistering speed, looking to outrun defenders at every opportunity.',
      fm_attributes: ['Pace', 'Acceleration', 'Off the Ball', 'Stamina'],
    },
    {
      playstyle: 'Tireless Shuttler',
      significance:
        'Runs the channel constantly, providing an outlet pass and stretching the opposition defense through sheer volume of running.',
      fm_attributes: ['Work Rate', 'Stamina', 'Off the Ball', 'Teamwork'],
    },
  ],
  AMC: [
    {
      playstyle: 'Creative Attacker',
      significance:
        "Operates in the 'hole' between midfield and defense, using creativity and vision to unlock the opposition defense with incisive passes.",
      fm_attributes: ['Passing', 'Vision', 'First Touch', 'Technique', 'Flair'],
    },
    {
      playstyle: 'Trequartista',
      significance:
        'A pure creator who operates in the final third with complete freedom, often absolved of defensive duties to conserve energy for attack.',
      fm_attributes: ['Flair', 'Vision', 'Passing', 'Technique', 'Off the Ball'],
    },
    {
      playstyle: 'Shadow Striker',
      significance:
        'A goalscoring attacking midfielder who makes aggressive, intelligent runs into the box to get on the end of chances, acting as a second striker.',
      fm_attributes: ['Finishing', 'Off the Ball', 'Anticipation', 'Pace', 'Composure'],
    },
    {
      playstyle: 'Enganche',
      significance:
        "A static playmaker who acts as the 'hook' linking midfield and attack. Does not roam, but serves as a pivot point for the team's attacks.",
      fm_attributes: ['Passing', 'Vision', 'First Touch', 'Technique', 'Decisions'],
    },
    {
      playstyle: 'Direct Attacker',
      significance:
        'A direct runner from deep who attacks the defensive line with dribbling and powerful shooting from range.',
      fm_attributes: ['Dribbling', 'Long Shots', 'Flair', 'Pace', 'Finishing'],
    },
    {
      playstyle: 'Pressing AMC',
      significance:
        "Leads the team's press from an advanced position, using energy and aggression to force turnovers high up the pitch.",
      fm_attributes: ['Work Rate', 'Aggression', 'Stamina', 'Anticipation', 'Tackling'],
    },
    {
      playstyle: 'Cannon',
      significance:
        'A player who specializes in shooting from distance, possessing immense power and a shoot-on-sight mentality.',
      fm_attributes: ['Long Shots', 'Finishing', 'Strength', 'Technique'],
    },
  ],
  AMR: [
    {
      playstyle: 'Inside Forward',
      significance:
        'Starts wide but cuts inside onto their stronger foot to shoot at goal, acting as a primary goalscoring threat.',
      fm_attributes: ['Finishing', 'Dribbling', 'Long Shots', 'Pace', 'Off the Ball'],
    },
    {
      playstyle: 'Inverted Winger',
      significance:
        'Similar to an Inside Forward, but focuses more on creating chances for others by cutting inside and playing through balls or combination passes.',
      fm_attributes: ['Passing', 'Vision', 'Dribbling', 'Technique', 'Composure'],
    },
    {
      playstyle: 'Raumdeuter (Space Investigator)',
      significance:
        'A unique role focused on finding and exploiting pockets of space. Drifts from a wide starting position to arrive in the box unmarked.',
      fm_attributes: ['Off the Ball', 'Anticipation', 'Finishing', 'Concentration', 'Decisions'],
    },
    {
      playstyle: 'Trickster',
      significance:
        'An unpredictable and skillful dribbler who uses flair and agility to beat defenders and create chaos.',
      fm_attributes: ['Flair', 'Dribbling', 'Agility', 'Acceleration', 'Technique'],
    },
    {
      playstyle: 'Wide Creator',
      significance:
        "Drifts inside from a wide starting position to act as the team's main creator, finding pockets of space to dictate play.",
      fm_attributes: ['Vision', 'Passing', 'Dribbling', 'Technique', 'Flair'],
    },
    {
      playstyle: 'Pressing Winger',
      significance:
        'Initiates the press from a wide area, using relentless energy to close down defenders and force turnovers in the attacking third.',
      fm_attributes: ['Work Rate', 'Aggression', 'Stamina', 'Acceleration'],
    },
    {
      playstyle: 'Traditional Winger (Attack)',
      significance:
        'A more attack-minded version of the traditional winger, focused on beating their man and getting into a position to shoot or create a clear chance.',
      fm_attributes: ['Dribbling', 'Pace', 'Flair', 'Crossing', 'Finishing'],
    },
  ],
  AML: [
    {
      playstyle: 'Inside Forward',
      significance:
        'Starts wide but cuts inside onto their stronger foot to shoot at goal, acting as a primary goalscoring threat.',
      fm_attributes: ['Finishing', 'Dribbling', 'Long Shots', 'Pace', 'Off the Ball'],
    },
    {
      playstyle: 'Inverted Winger',
      significance:
        'Similar to an Inside Forward, but focuses more on creating chances for others by cutting inside and playing through balls or combination passes.',
      fm_attributes: ['Passing', 'Vision', 'Dribbling', 'Technique', 'Composure'],
    },
    {
      playstyle: 'Raumdeuter (Space Investigator)',
      significance:
        'A unique role focused on finding and exploiting pockets of space. Drifts from a wide starting position to arrive in the box unmarked.',
      fm_attributes: ['Off the Ball', 'Anticipation', 'Finishing', 'Concentration', 'Decisions'],
    },
    {
      playstyle: 'Trickster',
      significance:
        'An unpredictable and skillful dribbler who uses flair and agility to beat defenders and create chaos.',
      fm_attributes: ['Flair', 'Dribbling', 'Agility', 'Acceleration', 'Technique'],
    },
    {
      playstyle: 'Wide Creator',
      significance:
        "Drifts inside from a wide starting position to act as the team's main creator, finding pockets of space to dictate play.",
      fm_attributes: ['Vision', 'Passing', 'Dribbling', 'Technique', 'Flair'],
    },
    {
      playstyle: 'Pressing Winger',
      significance:
        'Initiates the press from a wide area, using relentless energy to close down defenders and force turnovers in the attacking third.',
      fm_attributes: ['Work Rate', 'Aggression', 'Stamina', 'Acceleration'],
    },
    {
      playstyle: 'Traditional Winger (Attack)',
      significance:
        'A more attack-minded version of the traditional winger, focused on beating their man and getting into a position to shoot or create a clear chance.',
      fm_attributes: ['Dribbling', 'Pace', 'Flair', 'Crossing', 'Finishing'],
    },
  ],
  ST: [
    {
      playstyle: 'Poacher',
      significance:
        'A pure goalscorer who comes alive in the penalty box, relying on instinct and clinical finishing to convert chances.',
      fm_attributes: ['Finishing', 'Off the Ball', 'Anticipation', 'Composure', 'Acceleration'],
    },
    {
      playstyle: 'Bruiser',
      significance:
        'A physically imposing striker who acts as a focal point for the attack, holding up the ball, winning aerial duels, and bullying defenders.',
      fm_attributes: ['Strength', 'Jumping Reach', 'Heading', 'Balance', 'Aggression'],
    },
    {
      playstyle: 'Support Striker',
      significance:
        'Drops back towards the midfield to link play, creating space for wingers and attacking midfielders to run in behind.',
      fm_attributes: ['Passing', 'First Touch', 'Teamwork', 'Vision', 'Composure'],
    },
    {
      playstyle: 'Line Breaker',
      significance:
        'Leads the line by running in behind the defense, stretching the play, and acting as the primary goalscoring threat.',
      fm_attributes: ['Pace', 'Acceleration', 'Finishing', 'Off the Ball', 'Dribbling'],
    },
    {
      playstyle: 'Versatile Striker',
      significance:
        'A striker who possesses all the key attributes: can score, link play, create chances, and use both strength and pace.',
      fm_attributes: ['Finishing', 'Passing', 'Dribbling', 'Strength', 'Pace'],
    },
    {
      playstyle: 'Front Presser',
      significance:
        'The first line of defense. Harasses defenders relentlessly, forcing mistakes and winning the ball back high up the pitch.',
      fm_attributes: ['Work Rate', 'Aggression', 'Stamina', 'Bravery', 'Tackling'],
    },
    {
      playstyle: 'Withdrawn Striker',
      significance:
        'A striker who drops deep into midfield, dragging defenders out of position and creating space for teammates to exploit.',
      fm_attributes: ['Dribbling', 'Passing', 'Vision', 'Off the Ball', 'First Touch'],
    },
  ],
}
