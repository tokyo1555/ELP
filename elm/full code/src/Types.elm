module Types exposing
    ( Difficulte(..)
    , Etat(..)
    , Definition
    , Mode(..)
    , Model
    , Msg(..)
    , modeleInitial
    )


type Mode
    = Classique
    | Express


type Difficulte
    = Beginner
    | Medium
    | Expert


type Etat
    = Accueil
    | ChargementMots
    | ChoixMot
    | ChargementDefinitions
    | Pret
    | Gagne
    | TempsEcoule
    | Erreur String


type alias Definition =
    { typeMot : String
    , texte : String
    }


type alias Model =
    { etat : Etat
    , mode : Maybe Mode
    , difficulte : Maybe Difficulte
    , mots : List String
    , motSecret : String
    , definitionsVisibles : List Definition
    , saisie : String
    , message : String
    , tempsRestant : Int
    , tempsExpress : Int
    , score : Int
    }


modeleInitial : Model
modeleInitial =
    { etat = Accueil
    , mode = Nothing
    , difficulte = Nothing
    , mots = []
    , motSecret = ""
    , definitionsVisibles = []
    , saisie = ""
    , message = ""
    , tempsRestant = 0
    , tempsExpress = 500
    , score = 0
    }


type Msg
    = ChoisirDifficulte Difficulte
    | ChoisirMode Mode
    | TempsExpressChange String
    | LancerJeu
    | Home
    | Rejouer
    | Tick
    | MotsCharges String
    | MotChoisi String
    | DefinitionsChargees (List Definition)
    | SaisieChangee String
    | Verifier
    | MotSuivant
    | Echec String
    | PasserApresDelai
    | AfficherMot
