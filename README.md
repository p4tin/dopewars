# Dopewars

Classic Dopewars game written in pure GO.

### TODO

 1. Write Bank
 2. Write Loan Shark
    - 10% per day interest 
 3. Write Hospital
 4. Write random encounters
    - Stash
    - Bust
    - Craze
    - Market Flood
    
### Things in game

 5. Places
 
    Place|Business|Remarks
     --- | ------ | ----- 
    The Bronk|Loan Shark|A player can pay back his starting debt in this location.
    The Ghetto|Guns+Pockets|A player can purchase guns/pockets in this location.
    Central Park|Hospital|A player can be injured in a mugging or a police chase. If the player is injured they can be healed in a hospital.
    Manhattan|Bank|In a back the user can deposit or withdraw any amount of money. Money in a bank will be safe from muggings or the police.
    Brooklyn|Bank|In a back the user can deposit or withdraw any amount of money. Money in a bank will be safe from muggings or the police.
    Coney Island| | |
    
 6. Guns
 
    Name | Damage | Price
    ---- | ------ | -----
    Baretta | (Damage: 5, Accuracy: 50%) | $18,000 
    Ruger | (Damage: 4, Accuracy: 60%) | $14,500
    .38 Special | (Damage: 9, Accuracy: 50%) | $32,500
    Saturday Night Special | (Damage: 7, Accuracy: 65%) | $25,500

 7. Special Events

    Event | Description
    ----- | -----------
    Finding a Stash (Someone's Leftovers) | The player finds drugs for free. If the drugs exceed his carrying capacity, they are left behind.
    Mugging | If a player is mugged some number of the drugs he is carrying and some amount of the money he is carrying will be lost.
    Police Chase (Officer Hardass) | When chased by Officer Hardass, a player can choose to either run or stand and fight. If he runs there is a chance he will get away, as well as a chance to be injured. If he stands and fights, Hardass will fight back. If the player is killed he will lose the game. If Officer Hardass is killed the player will receive a monetary reward. Having a gun improved the players chances of winning against Hardass.
    Drug Busts | Occasionally, drug busts will occur. This will drive up the price of one specific type of drug. The drug affected and the amount by which the price is affected will be random.
    Addicts Craze | Occasionally, addicts will go nuts for a drug a drive up the prices. The effects are the same as a drug bust.
    Market Flood | Occasionally, additional shipments of a drug will occur. This will drive the price of that specific drug down massively. The specific amount and the drug affected are random.
    Bank Raid | Very rarely the police will find your bank account and seize all your assets.

  8. Final Score - The final score is calculated by adding your current amount of cash and your bank value. Then subtracting double the amount of your debt.

