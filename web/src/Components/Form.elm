module Components.Form exposing (ButtonStyle(..), DisabledVariant, button, input)

import Html as H exposing (Html)
import Html.Attributes as A
import Html.Events as E


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


type alias DisabledVariant =
    Bool


type
    -- TODO: those styles should get better naming
    ButtonStyle
    = Solid DisabledVariant
    | Bordered DisabledVariant
    | BorderedRedOnHover


button :
    { text : String
    , disabled : Bool
    , onClick : msg
    , style : ButtonStyle
    , type_ : String -- FIXME: move it into sep type
    }
    -> Html msg
button opts =
    H.button
        [ A.type_ opts.type_
        , A.class (buttonStyleToClass opts.style)
        , E.onClick opts.onClick
        , A.disabled opts.disabled
        ]
        [ H.text opts.text ]


buttonStyleToClass : ButtonStyle -> String
buttonStyleToClass style =
    case style of
        Solid isDisabled ->
            if isDisabled then
                "px-6 py-2 bg-gray-300 text-gray-500 rounded-md cursor-not-allowed transition-colors"

            else
                "px-6 py-2 bg-black text-white rounded-md hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"

        BorderedRedOnHover ->
            "text-gray-600 hover:text-red-600 transition-colors"

        Bordered isDisabled ->
            if isDisabled then
                "px-4 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors bg-green-100 border-green-300 text-green-700"

            else
                "px-4 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors border-gray-300 text-gray-700 hover:bg-gray-50"
