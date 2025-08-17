module Pages.Profile exposing (Model, Msg, ViewVariant, page)

import Api
import Api.Me
import Auth
import Components.Form
import Data.Me exposing (Me)
import Effect exposing (Effect)
import Html as H exposing (Html)
import Html.Attributes as A
import Layouts
import Page exposing (Page)
import Route exposing (Route)
import Shared
import Time.Format
import View exposing (View)


page : Auth.User -> Shared.Model -> Route () -> Page Model Msg
page _ shared _ =
    Page.new
        { init = init shared
        , update = update
        , subscriptions = subscriptions
        , view = view shared
        }
        |> Page.withLayout (\_ -> Layouts.Header {})



-- INIT


type alias Model =
    { view : ViewVariant
    , me : Api.Response Me
    }


init : Shared.Model -> () -> ( Model, Effect Msg )
init _ () =
    ( { view = Overview
      , me = Api.Loading
      }
    , Api.Me.get { onResponse = ApiMeResponded }
    )



-- UPDATE


type ViewVariant
    = Overview
    | Password
    | Email
    | DeleteAccount


type Msg
    = UserChangedView ViewVariant
    | ApiMeResponded (Result Api.Error Me)


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        UserChangedView variant ->
            ( { model | view = variant }, Effect.none )

        ApiMeResponded (Ok userData) ->
            ( { model | me = Api.Success userData }, Effect.none )

        ApiMeResponded (Err error) ->
            ( { model | me = Api.Failure error }, Effect.none )


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Shared.Model -> Model -> View Msg
view shared model =
    { title = "Profile"
    , body =
        [ H.div [ A.class "w-full max-w-4xl mx-auto" ]
            [ H.div [ A.class "bg-white rounded-lg border border-gray-200 shadow-sm" ]
                [ H.div [ A.class "p-6 pb-4 border-b border-gray-200" ]
                    [ H.h1 [ A.class "text-2xl font-bold text-gray-900" ] [ H.text "Account Settings" ]
                    , H.p [ A.class "text-gray-600 mt-2" ] [ H.text "Manage your account preferences and security settings" ]
                    ]
                , H.div [ A.class "flex" ]
                    [ viewNavigationSidebar model
                    , H.div [ A.class "flex-1 p-6" ]
                        [ case model.view of
                            Overview ->
                                viewProfileOverview shared model.me

                            Password ->
                                H.text "Password View"

                            Email ->
                                H.text "Email View"

                            DeleteAccount ->
                                H.text "Delete Account View"
                        ]
                    ]
                ]
            ]
        ]
    }


viewNavigationSidebar : Model -> Html Msg
viewNavigationSidebar model =
    let
        button : ViewVariant -> String -> Html Msg
        button variant text =
            Components.Form.button
                { text = text
                , onClick = UserChangedView variant
                , disabled = model.view == variant
                , style = Components.Form.PrimaryReverse (model.view == variant)
                }
    in
    H.div [ A.class "w-64 border-r border-gray-200 p-6" ]
        -- TODO: add icons to buttons
        [ H.div []
            [ H.nav [ A.class "[&>*]:w-full space-y-2" ]
                [ button Overview "Overview"
                , button Password "Password"
                , button Email "Email"
                , button DeleteAccount "Delete Account"
                ]
            ]
        ]


viewProfileOverview : Shared.Model -> Api.Response Me -> Html Msg
viewProfileOverview shared userResponse =
    case userResponse of
        Api.Success user ->
            H.div []
                [ H.h1 [] [ H.text "Profile Overview" ]
                , H.p [] [ H.text ("Created at: " ++ Time.Format.toString shared.timeZone user.createdAt) ]
                , H.p [] [ H.text ("Email: " ++ user.email) ]
                ]

        Api.Loading ->
            H.text "Loading..."

        Api.Failure err ->
            H.text (Api.errorMessage err)
