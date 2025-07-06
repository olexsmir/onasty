module Components.Utils exposing (loadSvg, viewIf, viewMaybe)

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


loadSvg : { path : String, class : String } -> Html msg
loadSvg { path, class } =
    H.img [ A.src ("/static/" ++ path), A.class class ] []
