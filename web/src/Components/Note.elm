module Components.Note exposing (noteIconSvg, noteNotFoundSvg)

import Svg exposing (Svg)
import Svg.Attributes as A


noteIconSvg : Svg msg
noteIconSvg =
    Svg.svg
        [ A.class "w-8 h-8 text-gray-400"
        , A.fill "none"
        , A.stroke "currentColor"
        , A.viewBox "0 0 24 24"
        ]
        [ Svg.path
            [ A.d "M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            , A.strokeWidth "2"
            , A.strokeLinecap "round"
            , A.strokeLinejoin "round"
            ]
            []
        ]


noteNotFoundSvg : Svg msg
noteNotFoundSvg =
    Svg.svg
        [ A.class "w-8 h-8 text-red-500"
        , A.fill "none"
        , A.stroke "currentColor"
        , A.viewBox "0 0 24 24"
        ]
        [ Svg.path
            [ A.d "M9.172 16.172a4 4 0 015.656 0M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            , A.strokeWidth "2"
            , A.strokeLinecap "round"
            , A.strokeLinejoin "round"
            ]
            []
        , Svg.path
            [ A.d "M6 18L18 6M6 6l12 12"
            , A.strokeWidth "2"
            , A.strokeLinecap "round"
            , A.strokeLinejoin "round"
            ]
            []
        ]
