const storage = {
    selectedStopIdsKey: "selectedStopIds",

    /**
     * Gets all selected stop IDs.
     * @return {string[]} selected IDs
     */
    getSelectedStopIds: function getSelectedStopIds() {
        const currentValue = localStorage.getItem(this.selectedStopIdsKey);
        /**
         * @type string[]
         */
        let selectedStopIds;
        if (currentValue == null) {
            selectedStopIds = [];
        } else {
            try {
                selectedStopIds = JSON.parse(currentValue);
            } catch(e) {
                selectedStopIds = [];
            }
        }
        return selectedStopIds;
    },
    /**
     * Gets if a stop is selected.
     * @param {string} stopId stop ID to check
     * @return {boolean} if it is selected
     */
    isSelected: function isSelected(stopId) {
        const selectedStopIds = this.getSelectedStopIds(stopId);
        return selectedStopIds.includes(stopId);
    },
    /**
     * Changes the selection state of a stop ID.
     * @param {string} stopId stop ID to change
     * @return {boolean} if it is selected after the switch
     */
    switchSelection: function switchSelection(stopId) {
        const selectedStopIds = this.getSelectedStopIds(stopId);
        const index = selectedStopIds.indexOf(stopId);
        const result = index < 0
        if (result) {
            selectedStopIds.push(stopId);
        } else {
            selectedStopIds.splice(index, 1);
        }
        localStorage.setItem(this.selectedStopIdsKey, JSON.stringify(selectedStopIds));
        return result;
    }
}
