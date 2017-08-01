[![Free music](http://img.auboutdufil.com/logo32.png)][Au Bout du Fil]

[![Build Status](https://travis-ci.org/Shywim/auboutdufil-api.svg?branch=master)](https://travis-ci.org/Shywim/auboutdufil-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/shywim/auboutdufil-api)](https://goreportcard.com/report/github.com/shywim/auboutdufil-api)
[![codecov](https://codecov.io/gh/Shywim/auboutdufil-api/branch/master/graph/badge.svg)](https://codecov.io/gh/Shywim/auboutdufil-api)

Au Bout du Fil unofficial API
=============================

Unofficial API endpoints for the [Au Bout du Fil] website using go.

A running server is available at https://auboutdufil.shywim.fr

## Usage

    make abdf
    ./abdf [-p port]

## Response format

The server answers with JSON data with the following model:

 - `title`: Title of the audio track
 - `artist`: Artist of the audio track
 - `genres`: Genres of the audio track as array
 - `license`: Audio track's license
 - `cover_art_url`: URL of the cover art
 - `download_url`: Download URL
 - `track_url`: Track's details url
 - `downloads`: Download count
 - `play_count`: Play count
 - `rating`: Users rating for the track
 - `published_date`: Date of publication on the website

## Endpoints

Each query give a response with (at most) 6 musics. You can add a `page` query param to get more musics, starting at 1.

### `/latest`

Provides the last musics published.

### `/best`

Provides the musics with the highest rating.

### `/downloads`

Provides the most downloaded musics.

### `/plays`

Provides the most played musics.

### Options

For each of theses endpoints you can add the following paths (in any order):

**License:**

    /license/:license

Filter by license. Those are the known licenses at this time (more may have been added to the site, you can pass them to this api):

 - Licence art libre (`ART-LIBRE`)
 - Creative Commons Attribution (`CC-BY`)
 - Creative Commons Attribution-Non Commercial (`CC-BYNC`)
 - Creative Commons Attribution-Non Commercial-No Derivative (`CC-BYNCND`)
 - Creative Commons Attribution-Non Commercial-Share Alike (`CC-BYNCSA`)
 - Creative Commons Attribution-No Derivative (`CC-BYND`)
 - Creative Commons Attribution-Share Alike (`CC-BYSA`)
 - Creative Commons Public Domain (`CC0`)

**Mood:**

    /mood/:mood

Filter by mood. Those are the known moods at this time (more may have been added to the site, you can pass them to this api):

 - `angry`
 - `bright`
 - `calm`
 - `dark`
 - `dramatic`
 - `funky`
 - `happy`
 - `inspirational`
 - `romantic`
 - `sad`

**Genre:**

    /genre/:genre

Filter by genre. Values passed will be used directly by the scrapper.

Genre, mood and license can also be passed as query params to the endpoint (e.g. `/latest?license=cc-by`). If you pass one of these both as a query param and as a url path, the url path will have priority over the query param.

For each of theses you can pass a value not listed here and the scrapper will try to get the music if there's any corresponding one.


 [Au Bout du Fil]: http://www.auboutdufil.com
