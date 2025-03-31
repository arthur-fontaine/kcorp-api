package leagueoflegends

// use
// http GET https://www.division2lol.fr/api/matches\?tournament_ids\=8414438806728540160 Range:"matches=0-127"
// to get the matches of division 2

// use
// Array.from(new Set(JSON.parse(document.querySelector('page-view-app').attributes.pages.value.slice(1, -1)).flatMap(b => b.settings?.context?.tournamentId || [])))
// or
// Array.from(new Set(Array.from(document.querySelector('page-view-app').attributes.pages.value.matchAll(/"tournamentId": ?"(\d+)"/g)).map(e => e[1])))
// in https://www.division2lol.fr/
// to get the tournament ids
