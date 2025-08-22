module Components.Box exposing (error, success, successBox, successText)

import Html as H exposing (Html)
import Html.Attributes as A


error : String -> Html msg
error errorMsg =
    H.div [ A.class "bg-red-50 border border-red-200 rounded-md p-4" ]
        [ H.p [ A.class "text-red-800 text-sm" ] [ H.text errorMsg ] ]


success : { header : String, body : String } -> Html msg
success opts =
    successBox
        [ H.div [ A.class "font-medium text-green-800 mb-2" ] [ H.text opts.header ]
        , H.p [ A.class "text-green-800 text-sm" ] [ H.text opts.body ]
        ]


successText : String -> Html msg
successText text =
    successBox [ H.p [ A.class "text-green-800 text-sm" ] [ H.text text ] ]


successBox : List (Html msg) -> Html msg
successBox child =
    H.div [ A.class "bg-green-50 border border-green-200 rounded-md p-4" ] child
