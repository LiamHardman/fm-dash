// src/services/playerService.js

const API_URL = "http://localhost:8091"; // Ensure this matches your Go backend port

export default {
  /**
   * Upload a player file to the API for parsing.
   * The API will store the data in-memory and return a datasetId and detectedCurrencySymbol.
   * @param {FormData} formData - The form data containing the file.
   * @returns {Promise<Object>} - A promise that resolves to an object
   * { datasetId: string, message: string, detectedCurrencySymbol: string }.
   */
  async uploadPlayerFile(formData) {
    try {
      const response = await fetch(`${API_URL}/upload`, {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(
          `API Error: ${response.status} - ${errorText || response.statusText}`,
        );
      }

      return await response.json(); // Expected: { datasetId, message, detectedCurrencySymbol }
    } catch (error) {
      console.error("Upload error:", error);
      throw error;
    }
  },

  /**
   * Fetches player data and the dataset's currency symbol from the backend using a dataset ID.
   * @param {string} datasetId - The ID of the dataset to retrieve.
   * @returns {Promise<Object>} - A promise that resolves to an object
   * { players: Player[], currencySymbol: string }.
   */
  async getPlayersByDatasetId(datasetId) {
    if (!datasetId) {
      return Promise.reject(new Error("Dataset ID is required."));
    }
    try {
      const response = await fetch(`${API_URL}/api/players/${datasetId}`);

      if (!response.ok) {
        if (response.status === 404) {
          throw new Error(
            `Player data not found for ID: ${datasetId}. The data might have expired or the ID is incorrect.`,
          );
        }
        const errorText = await response.text();
        throw new Error(
          `API Error: ${response.status} - ${errorText || response.statusText}`,
        );
      }
      // The backend now returns an object: { players: [], currencySymbol: "€" }
      return await response.json();
    } catch (error) {
      console.error("Error fetching players by dataset ID:", error);
      throw error;
    }
  },
};
