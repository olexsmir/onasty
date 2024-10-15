module Pages.HomePage exposing (view)

import Html exposing (Html, div, form, input, label, text)
import Model exposing (Model)


view : Model -> Html msg
view model =
    div []
        [ form []
            [ div []
                [ label [] [ text "Content" ]
                ]
            ]
        ]
