module Components.Error exposing (error)

import Html as H exposing (Html)
import Html.Attributes as A


error : String -> Html msg
error errorMsg =
    H.div [ A.class "bg-red-50 border border-red-200 rounded-md p-4" ]
        [ H.p [ A.class "text-red-800 text-sm" ] [ H.text errorMsg ] ]
