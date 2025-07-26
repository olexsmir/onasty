module Components.Utils exposing (commonContainer, viewIf, viewMaybe)

import Html as H exposing (Html)
import Html.Attributes as A


viewIf : Bool -> Html msg -> Html msg
viewIf condition html =
    if condition then
        html

    else
        H.text ""


viewMaybe : Maybe a -> (a -> Html msg) -> Html msg
viewMaybe maybeValue toHtml =
    case maybeValue of
        Just value ->
            toHtml value

        Nothing ->
            H.text ""


commonContainer : List (Html msg) -> Html msg
commonContainer child =
    H.div [ A.class "py-8 w-full max-w-4xl mx-auto " ]
        [ H.div [ A.class "rounded-lg border border-gray-200 shadow-sm" ] child ]
