module View exposing (view)

import Html exposing (Html, button, div, h1, input, li, span, text, ul)
import Html.Attributes as Attr exposing (placeholder, step, style, type_, value)
import Html.Events exposing (onClick, onInput)
import String
import Types exposing (Definition, Difficulte(..), Etat(..), Mode(..), Model, Msg(..))
import Html exposing (Html, a, button, div, h1, input, li, span, text, ul)



view : Model -> Html Msg
view model =
    div
        [ style "min-height" "100vh"
        , style "background" "#0b1020"
        , style "color" "#e8ecff"
        , style "font-family" "system-ui, -apple-system, Segoe UI, Roboto, Arial"
        , style "padding" "32px"
        ]
        [ div
            [ style "max-width" "920px"
            , style "margin" "0 auto"
            ]
            [ barreHaut model
            , carte (vueSelonEtat model)
            ]
        ]


vide : Html Msg
vide =
    text ""


carte : Html Msg -> Html Msg
carte contenu =
    div
        [ style "background" "rgba(255,255,255,0.06)"
        , style "border" "1px solid rgba(255,255,255,0.10)"
        , style "border-radius" "16px"
        , style "padding" "18px"
        , style "box-shadow" "0 10px 30px rgba(0,0,0,0.35)"
        ]
        [ contenu ]


vueSelonEtat : Model -> Html Msg
vueSelonEtat model =
    case model.etat of
        Accueil ->
            vueAccueil model

        ChargementMots ->
            badgeInfo "Chargement des mots..."

        ChoixMot ->
            badgeInfo "Choix du mot..."

        ChargementDefinitions ->
            badgeInfo "Chargement des définitions..."

        Pret ->
            vueJeu model

        Gagne ->
            vueGagne model

        TempsEcoule ->
            vueTempsEcoule model

        Erreur msgErr ->
            badgeErreur ("Erreur : " ++ msgErr)



-- HEADER


barreHaut : Model -> Html Msg
barreHaut model =
    div
        [ style "display" "flex"
        , style "justify-content" "space-between"
        , style "align-items" "flex-start"
        , style "gap" "12px"
        , style "margin-bottom" "18px"
        ]
    [ entete
    , div
        [ style "display" "flex"
        , style "gap" "8px"
        , style "align-items" "center"
        ]
        ([ boutonLienCompact "📄 README" "README.md" ]
            ++ (if model.etat == Accueil then
                    []
                else
                    [ boutonHome ]
            )
        )
    ]

entete : Html Msg
entete =
    div []
        [ h1
            [ style "margin" "0"
            , style "font-size" "34px"
            , style "letter-spacing" "0.2px"
            ]
            [ text "GuessIt" ]
        , div
            [ style "opacity" "0.85"
            , style "margin-top" "6px"
            ]
            [ text "Bienvenue dans GuessIt, votre jeu préféré, le but du jeu est de deviner le mot secret grâce à des définitions. \n Choisis un niveau de difficulté (Beginner, Medium, Expert) et le mode de jeu (Classique, Express)." ]
        ]


boutonHome : Html Msg
boutonHome =
    button
        [ onClick Home
        , style "padding" "10px 14px"
        , style "border-radius" "12px"
        , style "border" "1px solid rgba(255,255,255,0.18)"
        , style "background" "rgba(255,255,255,0.08)"
        , style "color" "#e8ecff"
        , style "cursor" "pointer"
        , style "font-weight" "700"
        ]
        [ text "Menu" ]



-- ACCUEIL

boutonLien : String -> String -> Html Msg
boutonLien label url =
    a
        [ Attr.href url
        , Attr.target "_blank"
        , style "display" "inline-block"
        , style "padding" "10px 14px"
        , style "border-radius" "12px"
        , style "border" "1px solid rgba(255,255,255,0.18)"
        , style "background" "rgba(255,255,255,0.08)"
        , style "color" "#e8ecff"
        , style "cursor" "pointer"
        , style "font-weight" "800"
        , style "text-decoration" "none"
        ]
        [ text label ]

vueAccueil : Model -> Html Msg
vueAccueil model =

    let
        pret =
            model.mode /= Nothing && model.difficulte /= Nothing
    in
    div []
        [ sectionTitre "Difficulté 🎚️"
        , div [ style "display" "flex"
    , style "gap" "16px"
    , style "flex-wrap" "wrap"
    , style "margin-top" "15px"
    ]
        [ boutonOption (model.difficulte == Just Beginner) "Beginner 🌱" (ChoisirDifficulte Beginner)
        , boutonOption (model.difficulte == Just Medium) "Medium ⚡" (ChoisirDifficulte Medium)
        , boutonOption (model.difficulte == Just Expert) "Expert 🔥" (ChoisirDifficulte Expert)
    ]
        , separateur
        , sectionTitre "Mode 🎮"
        , div     [ style "display" "flex"
    , style "gap" "16px"
    , style "flex-wrap" "wrap"
    , style "margin-top" "14px"
    ]
        [ boutonOption (model.mode == Just Classique) "Classique 🎯" (ChoisirMode Classique)
        , boutonOption (model.mode == Just Express) "Express ⏱️" (ChoisirMode Express)
        ]

        , case model.mode of
            Just Express ->
                div [ style "margin-top" "10px" ]
                    [ text ("Temps : " ++ String.fromInt model.tempsExpress ++ " s")
                    , input
                        [ type_ "range"
                        , Attr.min "60"
                        , Attr.max "500"
                        , step "20"
                        , value (String.fromInt model.tempsExpress)
                        , onInput TempsExpressChange
                        , style "width" "100%"
                        ]
                        []
                    ]

            _ ->
                vide
        , div [ style "margin-top" "14px" ]
            [ boutonJouer pret ]
        , if model.message /= "" then
            messageBox model.message False
          else
            vide
        ]


boutonJouer : Bool -> Html Msg
boutonJouer pret =
    button
        ([ onClick LancerJeu
         , style "padding" "10px 14px"
         , style "border-radius" "12px"
         , style "border" "1px solid rgba(255,255,255,0.12)"
         , style "background" "#facd5b"
         , style "color" "#000000"
         , style "cursor" "pointer"
         , style "font-weight" "700"
         , style "margin-top" "8px"
         ]
            ++ (if pret then
                    []
                else
                    [ Attr.disabled True
                    , style "opacity" "0.5"
                    , style "cursor" "not-allowed"
                    ]
               )
        )
        [ text "Lancer le jeu" ]



boutonOption : Bool -> String -> Msg -> Html Msg
boutonOption selected label msg =
    button
        [ onClick msg
        , style "padding" "10px 14px"
        , style "border-radius" "12px"
        , style "cursor" "pointer"
        , style "font-weight" "800"
        , style "border"
            (if selected then
                "1px solid rgba(250,205,91,0.9)"
             else
                "1px solid rgba(255,255,255,0.18)"
            )
        , style "background"
            (if selected then
                "#facd5b"
             else
                "rgba(255,255,255,0.08)"
            )
        , style "color"
            (if selected then
                "#000000"
             else
                "#e8ecff"
            )
        ]
        [ text label ]



-- JEU


vueJeu : Model -> Html Msg
vueJeu model =
    div []
        [ bandeauExpress model
        , sectionTitre "Définitions"
        , listeDefinitions model.definitionsVisibles
        , separateur
        , sectionTitre "Ta réponse"
        , zoneReponse model
        , if model.message /= "" then
            messageBox model.message False
          else
            vide
        ]


bandeauExpress : Model -> Html Msg
bandeauExpress model =
    case model.mode of
        Just Express ->
            div
                [ style "display" "flex"
                , style "gap" "10px"
                , style "flex-wrap" "wrap"
                , style "align-items" "center"
                , style "margin" "8px 0 10px 0"
                ]
                [ badgeChiffre "Temps" (String.fromInt model.tempsRestant ++ "s")
                , badgeChiffre "Score" (String.fromInt model.score)
                ]

        _ ->
            vide


zoneReponse : Model -> Html Msg
zoneReponse model =
    div
        [ style "display" "flex"
        , style "gap" "10px"
        , style "align-items" "center"
        , style "flex-wrap" "wrap"
        , style "margin-top" "10px"
        ]
        ([ input
            [ value model.saisie
            , placeholder "Tape le mot"
            , onInput SaisieChangee
            , style "flex" "1"
            , style "min-width" "220px"
            , style "padding" "10px 12px"
            , style "border-radius" "12px"
            , style "border" "1px solid rgba(255,255,255,0.18)"
            , style "background" "rgba(0,0,0,0.25)"
            , style "color" "#e8ecff"
            , style "outline" "none"
            ]
            []
         , boutonPrincipal "Vérifier" Verifier
         ]
            ++ (case model.mode of
                    Just Classique ->
                        [ boutonSecondaire "Afficher le mot" AfficherMot
                        , boutonSecondaire "Mot suivant" MotSuivant
                        ]

                    Just Express ->
                        [ boutonSecondaire "Mot suivant" MotSuivant ]

                    Nothing ->
                        [ boutonSecondaire "Mot suivant" MotSuivant ]
               )
        )


vueGagne : Model -> Html Msg
vueGagne model =
    div []
        [ sectionTitre "Bravo"
        , messageBox model.message True
        , boutonPrincipal "Mot suivant" MotSuivant
        ]


vueTempsEcoule : Model -> Html Msg
vueTempsEcoule model =
    div []
        [ sectionTitre "Temps écoulé"
        , messageBox ("Score : " ++ String.fromInt model.score) False
        , boutonPrincipal "Rejouer" Rejouer
        ]



-- DEFINITIONS


listeDefinitions : List Definition -> Html Msg
listeDefinitions defs =
    if List.isEmpty defs then
        div [ style "opacity" "0.85" ] [ text "Aucune définition trouvée." ]

    else
        ul [ style "padding-left" "18px", style "line-height" "1.6" ]
            (List.map (\d -> li [ style "margin" "6px 0" ] [ text (d.typeMot ++ " : " ++ d.texte) ]) defs)



-- UI


messageBox : String -> Bool -> Html Msg
messageBox msg estGagne =
    div
        [ style "margin-top" "14px"
        , style "padding" "12px 12px"
        , style "border-radius" "12px"
        , style "border" "1px solid rgba(255,255,255,0.14)"
        , style "background"
            (if estGagne then
                "rgba(90, 200, 120, 0.20)"
             else
                "rgba(255, 60, 60, 0.18)"
            )
        ]
        [ text msg ]


badgeInfo : String -> Html Msg
badgeInfo contenu =
    div [ style "opacity" "0.9" ] [ text contenu ]


badgeErreur : String -> Html Msg
badgeErreur contenu =
    div [ style "opacity" "0.9" ] [ text contenu ]


badgeChiffre : String -> String -> Html Msg
badgeChiffre titreTxt valeurTxt =
    div
        [ style "display" "inline-flex"
        , style "gap" "8px"
        , style "align-items" "center"
        , style "padding" "6px 10px"
        , style "border-radius" "999px"
        , style "background" "rgba(255,255,255,0.08)"
        , style "border" "1px solid rgba(255,255,255,0.12)"
        , style "font-size" "16px"
        ]
        [ span [ style "opacity" "0.75" ] [ text titreTxt ]
        , span [ style "font-weight" "800" ] [ text valeurTxt ]
        ]


sectionTitre : String -> Html Msg
sectionTitre t =
    div
        [ style "margin-top" "6px"
        , style "font-weight" "800"
        , style "opacity" "0.95"
        ]
        [ text t ]


separateur : Html Msg
separateur =
    div
        [ style "height" "1px"
        , style "background" "rgba(255,255,255,0.12)"
        , style "margin" "16px 0"
        ]
        []

boutonLienCompact : String -> String -> Html Msg
boutonLienCompact label url =
    a
        [ Attr.href url
        , Attr.target "_blank"
        , style "padding" "10px 20px"
        , style "border-radius" "10px"
        , style "border" "1px solid rgba(255,255,255,0.18)"
        , style "background" "rgba(255,255,255,0.06)"
        , style "color" "#e8ecff"
        , style "cursor" "pointer"
        , style "font-weight" "700"
        , style "font-size" "14px"
        , style "text-decoration" "none"
        , style "text-align" "center"
        ]
        [ text label ]

boutonPrincipal : String -> Msg -> Html Msg
boutonPrincipal label msg =
    button
        [ onClick msg
        , style "padding" "10px 14px"
        , style "border-radius" "12px"
        , style "border" "1px solid rgba(255,255,255,0.12)"
        , style "background" "#facd5b"
        , style "color" "#000000"
        , style "cursor" "pointer"
        , style "font-weight" "800"
        ]
        [ text label ]


boutonSecondaire : String -> Msg -> Html Msg
boutonSecondaire label msg =
    button
        [ onClick msg
        , style "padding" "10px 14px"
        , style "border-radius" "12px"
        , style "border" "1px solid rgba(255,255,255,0.18)"
        , style "background" "rgba(255,255,255,0.08)"
        , style "color" "#e8ecff"
        , style "cursor" "pointer"
        , style "font-weight" "800"
        ]
        [ text label ]

