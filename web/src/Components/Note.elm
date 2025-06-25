module Components.Note exposing (noteIconSvg)

import Svg exposing (Svg)
import Svg.Attributes


noteIconSvg : Svg msg
noteIconSvg =
    Svg.svg
        [ Svg.Attributes.class "w-8 h-8 text-gray-400"
        , Svg.Attributes.fill "none"
        , Svg.Attributes.stroke "currentColor"
        , Svg.Attributes.viewBox "0 0 24 24"
        ]
        [ Svg.path
            [ Svg.Attributes.d "M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            , Svg.Attributes.strokeWidth "2"
            , Svg.Attributes.strokeLinecap "round"
            , Svg.Attributes.strokeLinejoin "round"
            ]
            []
        ]
