const api = {
    /**
     * Fetches upcoming.
     * @param {string[]} stopIds 
     * @param {string} until YYYYMMDD hh:mm:ss format date
     * @return {Promise<{id: string; stopName: string; departures: {departureTime: string, route: { name: string; headSign: string; color: string; textColor: string}}[]}[]>}
     */
    fetchUpcoming: async function fetchUpcoming(stopIds, until) {
        const upcomingUrl = new URL(location.origin + "/api/upcoming");
        stopIds.forEach(stopId => upcomingUrl.searchParams.append("stopId", stopId));
        if (until != null) {
            upcomingUrl.searchParams.append("until", until);
        }
        const res = await fetch(upcomingUrl);
        if (res.ok) {
            return res.json();
        }
        const error = await res.text();
        throw new Error(`${res.status} ${res.statusText} ${error}`);
    },
    /**
     * Fetches stops.
     * @param {string stopName name part 
     * @return {Promise<{id: string; stopName: string; routes: { name: string; headSign: string; color: string; textColor: string}[]}[]>}
     */
    fetchStops: async function fetchStops(stopName) {
        const stopsUrl = new URL(location.origin + "/api/findstop");
        stopsUrl.searchParams.append("stopName", stopName);
        const res = await fetch(stopsUrl);
        if (res.ok) {
            return res.json();
        }
        const error = await res.text();
        throw new Error(`${res.status} ${res.statusText} ${error}`);
    }
};