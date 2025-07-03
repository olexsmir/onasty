module Components.Form exposing (input)

import Html as H exposing (Html)
import Html.Attributes as A
import Html.Events as E


input :
    { id : String
    , field : field
    , label : String
    , type_ : String
    , value : String
    , placeholder : String
    , required : Bool
    , helpText : Maybe String
    , prefix : Maybe String

    -- , error : Maybe String  -- show user the error
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
