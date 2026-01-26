module Words exposing (chargerMots, parserMots, choisirMot)

import Http
import Random
import String
import Types exposing (Msg(..))


chargerMots : Cmd Msg
chargerMots =
    Http.get
        { url = "data/Words.txt"
        , expect =
            Http.expectString
                (\result ->
                    case result of
                        Ok contenu ->
                            MotsCharges contenu

                        Err _ ->
                            Echec "Impossible de charger Words.txt"
                )
        }


parserMots : String -> List String
parserMots contenu =
    contenu
        |> String.lines
        |> List.map String.trim
        |> List.filter (\l -> l /= "")


choisirMot : List String -> Cmd Msg
choisirMot mots =
    case mots of
        [] ->
            Random.generate MotChoisi (Random.constant "")

        x :: xs ->
            Random.generate MotChoisi (Random.uniform x xs)
