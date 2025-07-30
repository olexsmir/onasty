module Components.Form exposing (ButtonStyle(..), CanBeClicked, button, input, submitButton)

import Html as H exposing (Html)
import Html.Attributes as A
import Html.Events as E



-- INPUT


input :
    -- TODO: add `error : Maybe String`, to show that field is not correct and message
    { id : String
    , field : field
    , label : String
    , type_ : String
    , value : String
    , placeholder : String
    , required : Bool
    , helpText : Maybe String
    , prefix : Maybe String
    , onInput : String -> msg
    }
    -> Html msg
input opts =
    H.div [ A.class "space-y-2" ]
        [ H.label
            [ A.for opts.id
            , A.class "block text-sm font-medium text-gray-700"
            ]
            [ H.text opts.label ]
        , H.div
            [ A.class
                (if opts.prefix /= Nothing then
                    "flex items-center"

                 else
                    ""
                )
            ]
            [ case opts.prefix of
                Just prefix ->
                    H.span [ A.class "text-gray-500 text-md mr-2 whitespace-nowrap" ] [ H.text prefix ]

                Nothing ->
                    H.text ""
            , H.input
                [ A.class "w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent"
                , A.type_ opts.type_
                , A.value opts.value
                , A.id opts.id
                , A.placeholder opts.placeholder
                , A.required opts.required
                , E.onInput opts.onInput
                ]
                []
            ]
        , case opts.helpText of
            Just help ->
                H.p [ A.class "text-xs text-gray-500 mt-1" ] [ H.text help ]

            Nothing ->
                H.text ""
        ]



-- BUTTON


type alias CanBeClicked =
    Bool


type ButtonStyle
    = Primary CanBeClicked
    | Secondary CanBeClicked
    | SecondaryDisabled CanBeClicked
    | SecondaryDanger


button : { text : String, disabled : Bool, onClick : msg, style : ButtonStyle } -> Html msg
button opts =
    H.button
        [ A.type_ "button"
        , E.onClick opts.onClick
        , A.class (buttonStyleToClass opts.style "")
        , A.disabled opts.disabled
        ]
        [ H.text opts.text ]


submitButton : { text : String, disabled : Bool, class : String, style : ButtonStyle } -> Html msg
submitButton opts =
    H.button
        [ A.type_ "submit"
        , A.class (buttonStyleToClass opts.style opts.class)
        , A.disabled opts.disabled
        ]
        [ H.text opts.text ]


buttonStyleToClass : ButtonStyle -> String -> String
buttonStyleToClass style appendClasses =
    case style of
        Primary canBeClicked ->
            getButtonClasses canBeClicked
                appendClasses
                "px-6 py-2 bg-gray-300 text-gray-500 rounded-md cursor-not-allowed transition-colors"
                "px-6 py-2 bg-black text-white rounded-md hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"

        SecondaryDanger ->
            "text-gray-600 hover:text-red-600 transition-colors"

        Secondary canBeClicked ->
            getButtonClasses canBeClicked
                appendClasses
                "px-4 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors bg-green-100 border-green-300 text-green-700"
                "px-4 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors border-gray-300 text-gray-700 hover:bg-gray-50"

        SecondaryDisabled canBeClicked ->
            getButtonClasses canBeClicked
                appendClasses
                "w-full px-4 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors mt-3 border border-gray-300 text-gray-400 cursor-not-allowed"
                "w-full px-4 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors mt-3 border border-gray-300 text-gray-700 hover:bg-gray-50"


getButtonClasses : Bool -> String -> String -> String -> String
getButtonClasses cond extend whenTrue whenFalse =
    let
        cls =
            if String.isEmpty extend then
                ""

            else
                " " ++ extend
    in
    if cond then
        whenTrue ++ cls

    else
        whenFalse ++ cls
