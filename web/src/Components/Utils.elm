module Components.Utils exposing (viewIf, viewMaybe)

import Html exposing (Html)


viewIf : Bool -> Html msg -> Html msg
viewIf condition html =
    if condition then
        html

    else
        Html.text ""


viewMaybe : Maybe a -> (a -> Html msg) -> Html msg
viewMaybe maybeValue toHtml =
    case maybeValue of
        Just value ->
            toHtml value

        Nothing ->
            Html.text ""
