module Main exposing (main)

import Browser
import Http
import Json.Decode as D
import String
import Time
import Types exposing (..)
import View
import Words
import Process
import Task

main : Program () Model Msg
main =
    Browser.element
        { init = \_ -> ( Types.modeleInitial, Cmd.none )
        , update = update
        , view = View.view
        , subscriptions = subscriptions
        }


subscriptions : Model -> Sub Msg
subscriptions model =
    case ( model.mode, model.etat ) of
        ( Just Express, Pret ) ->
            Time.every 1000 (\_ -> Tick)

        ( Just Express, ChargementMots ) ->
            Time.every 1000 (\_ -> Tick)

        ( Just Express, ChoixMot ) ->
            Time.every 1000 (\_ -> Tick)

        ( Just Express, ChargementDefinitions ) ->
            Time.every 1000 (\_ -> Tick)

        _ ->
            Sub.none


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        ChoisirDifficulte d ->
            ( { model | difficulte = Just d, message = "" }
            , Cmd.none
            )

        ChoisirMode m ->
            let
                m2 =
                    { model | mode = Just m, message = "" }

                m3 =
                    case m of
                        Express ->
                            { m2 | tempsRestant = m2.tempsExpress }

                        Classique ->
                            { m2 | tempsRestant = 0 }
            in
            ( m3, Cmd.none )


        TempsExpressChange v ->
            case String.toInt v of
                Just t ->
                    ( { model | tempsExpress = t, tempsRestant = t }, Cmd.none )

                Nothing ->
                    ( model, Cmd.none )

        LancerJeu ->
            if model.difficulte == Nothing then
                ( { model | message = "Choisis une difficulté." }, Cmd.none )

            else if model.mode == Nothing then
                ( { model | message = "Choisis un mode." }, Cmd.none )

            else
                lancer model

        Home ->
            ( Types.modeleInitial, Cmd.none )

        Rejouer ->
            ( Types.modeleInitial, Cmd.none )

        Tick ->
            case model.mode of
                Just Express ->
                    if model.tempsRestant <= 1 then
                        ( { model | tempsRestant = 0, etat = TempsEcoule }, Cmd.none )

                    else
                        ( { model | tempsRestant = model.tempsRestant - 1 }, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        MotsCharges contenu ->
            let
                liste =
                    Words.parserMots contenu
            in
            if List.isEmpty liste then
                ( { model | etat = Erreur "Words.txt vide." }, Cmd.none )

            else
                ( { model | mots = liste, etat = ChoixMot }
                , Words.choisirMot liste
                )

        MotChoisi mot ->
            ( { model
                | motSecret = mot
                , etat = ChargementDefinitions
                , definitionsVisibles = []
                , saisie = ""
                , message = ""
              }
            , chargerDefinitions mot
            )

        DefinitionsChargees defs ->
            let
                dif =
                    Maybe.withDefault Beginner model.difficulte

                visibles =
                    defsSelonDifficulte dif defs
            in
            ( { model | definitionsVisibles = visibles, etat = Pret }, Cmd.none )

        SaisieChangee txt ->
            ( { model | saisie = txt }, Cmd.none )

        Verifier ->
            let
                entree =
                    String.toLower (String.trim model.saisie)

                cible =
                    String.toLower (String.trim model.motSecret)
            in
            if entree == cible then
                ( { model | score = model.score + 1, message = "Vous avez trouvez !" }
                , Cmd.none
                )

            else
                case model.mode of
                    Just Express ->
                        ( { model
                            | score = max 0 (model.score - 1)
                            , message = "Faux. La bonne réponse était : " ++ model.motSecret
                          }
                        , Task.perform (\_ -> PasserApresDelai) (Process.sleep 1200)
                        )

                    _ ->
                        ( { model | message = "Faux, réessaie." }
                        , Cmd.none
                        )


        PasserApresDelai ->
            ( { model
                | definitionsVisibles = []
                , etat = ChoixMot
                , message = ""
                , motSecret = ""
                , saisie = ""
              }
            , Words.choisirMot model.mots
            )


        MotSuivant ->
            ( { model
                | etat = ChoixMot
                , motSecret = ""
                , definitionsVisibles = []
                , saisie = ""
                , message = ""
              }
            , Words.choisirMot model.mots
            )
   
        AfficherMot ->
            ( { model | message = "Le mot était : " ++ model.motSecret }
            , Cmd.none
            )

        Echec err ->
            ( { model | etat = Erreur err }, Cmd.none )



lancer : Model -> ( Model, Cmd Msg )
lancer model =
    let
        tRest =
            case model.mode of
                Just Express ->
                    model.tempsExpress

                _ ->
                    0
    in
    ( { model
        | etat = ChargementMots
        , message = ""
        , score = 0
        , tempsRestant = tRest
        , mots = []
        , motSecret = ""
        , definitionsVisibles = []
        , saisie = ""
      }
    , Words.chargerMots
    )


defsSelonDifficulte : Difficulte -> List Definition -> List Definition
defsSelonDifficulte dif defs =
    case dif of
        Beginner ->
            defs

        Medium ->
            defs |> List.take 2

        Expert ->
            defs |> List.take 1



-- HTTP

chargerDefinitions : String -> Cmd Msg
chargerDefinitions mot =
    Http.get
        { url = "https://api.dictionaryapi.dev/api/v2/entries/en/" ++ mot
        , expect =
            Http.expectJson
                (\res ->
                    case res of
                        Ok defs ->
                            DefinitionsChargees defs

                        Err _ ->
                            Echec "Impossible de récupérer les définitions."
                )
                decoderDefinitions
        }


decoderDefinitions : D.Decoder (List Definition)
decoderDefinitions =
    D.list decoderEntree |> D.map List.concat


decoderEntree : D.Decoder (List Definition)
decoderEntree =
    D.field "meanings" (D.list decoderMeaning) |> D.map List.concat


decoderMeaning : D.Decoder (List Definition)
decoderMeaning =
    D.map2
        (\pos defs -> List.map (\d -> { typeMot = pos, texte = d }) defs)
        (D.field "partOfSpeech" D.string)
        (D.field "definitions" (D.list (D.field "definition" D.string)))



normaliser : String -> String
normaliser s =
    s |> String.trim |> String.toLower

