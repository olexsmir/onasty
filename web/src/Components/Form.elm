module Components.Form exposing (input)

import Html as H exposing (Html)
import Html.Attributes as A
import Html.Events as E


input :
    { field : field
    , label : String
    , type_ : String
    , value : String
    , placeholder : String
    , required : Bool
    , onInput : String -> msg
    }
    -> Html msg
input opts =
    H.div [ A.class "space-y-2" ]
        [ H.label
            [ A.class "block text-sm font-medium text-gray-700" ]
            [ H.text opts.label ]
        , H.div [] [ H.input
                [ A.class "w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent"
                , A.type_ opts.type_
                , A.value opts.value
                , A.placeholder opts.placeholder
                , A.required opts.required
                , E.onInput opts.onInput
                ]
                []
            ]
        ]
