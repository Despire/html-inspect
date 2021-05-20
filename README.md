# HTML INSPECT

Extract general information about the HTML page.

## How to run

```
    make run
```

## Example

In terminal 1
```
    make run
```

In terminal 2
```
    curl -X POST -d '{"url": "https://www.facebook.com"}' http://127.0.0.1:8080 -H "Content-Type: application/json"


    ...

    {
    "version": "5",
    "title": "Facebook – prihláste sa alebo sa zaregistrujte",
    "login_form": true,
    "headings": [
        {
            "level": "h2",
            "total": 1
        }
    ],
    "internal": {
        "domain": "www.facebook.com",
        "links": [
            "https://www.facebook.com/directory/people/",
            "https://www.facebook.com/marketplace/",
            "https://www.facebook.com/policies/cookies/",
            "https://www.facebook.com/games/",
            "https://www.facebook.com/login/",
            "https://www.facebook.com/places/",
            "https://www.facebook.com/jobs/",
            "https://www.facebook.com/fundraisers/",
            "https://www.facebook.com/pages/create/?ref_type=site_footer",
            "https://www.facebook.com/privacy/explanation",
            "https://www.facebook.com/pages/create/?ref_type=registration_form",
            "https://www.facebook.com/r.php",
            "https://www.facebook.com/lite/",
            "https://www.facebook.com/local/lists/245019872666104/",
            "https://www.facebook.com/policies?ref=pf",
            "https://www.facebook.com/directory/pages/",
            "https://www.facebook.com/votinginformationcenter/?entry_point=c2l0ZQ%3D%3D",
            "https://www.facebook.com/allactivity?privacy_source=activity_log_top_menu",
            "https://www.facebook.com#",
            "https://www.facebook.com/directory/places/",
            "https://www.facebook.com/careers/?ref=pf",
            "https://www.facebook.com/help/?ref=pf",
            "https://www.facebook.com/ad_campaign/landing.php?placement=pflo\u0026campaign_id=402047449186\u0026nav_source=unknown\u0026extra_1=auto",
            "https://www.facebook.com/directory/groups/",
            "https://www.facebook.com/biz/directory/",
            "https://www.facebook.com/pages/category/",
            "https://www.facebook.com/settings",
            "https://www.facebook.com/recover/initiate/?ars=facebook_login\u0026privacy_mutation_token=eyJ0eXBlIjowLCJjcmVhdGlvbl90aW1lIjoxNjIxNTE3MzgxLCJjYWxsc2l0ZV9pZCI6MzgxMjI5MDc5NTc1OTQ2fQ%3D%3D",
            "https://www.facebook.com/",
            "https://www.facebook.com/watch/",
            "https://www.facebook.com/help/568137493302217"
        ],
        "total": 31
    },
    "external": [
        {
            "domain": "hu-hu.facebook.com",
            "links": [
                "https://hu-hu.facebook.com/"
            ],
            "total": 1
        },
        {
            "domain": "messenger.com",
            "links": [
                "https://messenger.com/"
            ],
            "total": 1
        },
        {
            "domain": "it-it.facebook.com",
            "links": [
                "https://it-it.facebook.com/"
            ],
            "total": 1
        },
        {
            "domain": "pay.facebook.com",
            "links": [
                "https://pay.facebook.com/"
            ],
            "total": 1
        },
        {
            "domain": "www.oculus.com",
            "links": [
                "https://www.oculus.com/"
            ],
            "total": 1
        },
        {
            "domain": "about.facebook.com",
            "links": [
                "https://about.facebook.com/"
            ],
            "total": 1
        },
        {
            "domain": "developers.facebook.com",
            "links": [
                "https://developers.facebook.com/?ref=pf"
            ],
            "total": 1
        },
        {
            "domain": "de-de.facebook.com",
            "links": [
                "https://de-de.facebook.com/"
            ],
            "total": 1
        },
        {
            "domain": "es-la.facebook.com",
            "links": [
                "https://es-la.facebook.com/"
            ],
            "total": 1
        },
        {
            "domain": "pt-br.facebook.com",
            "links": [
                "https://pt-br.facebook.com/"
            ],
            "total": 1
        },
        {
            "domain": "portal.facebook.com",
            "links": [
                "https://portal.facebook.com/"
            ],
            "total": 1
        },
        {
            "domain": "l.facebook.com",
            "links": [
                "https://l.facebook.com/l.php?u=https%3A%2F%2Fwww.instagram.com%2F\u0026h=AT33KN8xyDocZ69aW_aW-bzhLJzQbADZ5iVfecEEAoXBAdRnDrBsi62D_5zyZm8N03CehLGkqnnMCgi2BnamEkXZRam09JAZfW6PpC_okPMg4gVHCPEE_nJXyU7ruky5HYKjK6DIVK2i303jEltTfg2lNZNB7w"
            ],
            "total": 1
        },
        {
            "domain": "cs-cz.facebook.com",
            "links": [
                "https://cs-cz.facebook.com/"
            ],
            "total": 1
        },
        {
            "domain": "ru-ru.facebook.com",
            "links": [
                "https://ru-ru.facebook.com/"
            ],
            "total": 1
        },
        {
            "domain": "fr-fr.facebook.com",
            "links": [
                "https://fr-fr.facebook.com/"
            ],
            "total": 1
        },
        {
            "domain": "vi-vn.facebook.com",
            "links": [
                "https://vi-vn.facebook.com/"
            ],
            "total": 1
        }
    ],
    "inaccessible": [
        {
            "domain": "www.facebook.com",
            "links": [
                {
                    "URL": "https://www.facebook.com/pages/create/?ref_type=site_footer",
                    "Reason": "endpoint responded with code: 500"
                },
                {
                    "URL": "https://www.facebook.com/pages/create/?ref_type=registration_form",
                    "Reason": "endpoint responded with code: 500"
                }
            ],
            "total": 2
        }
    ]
}
```
