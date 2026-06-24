/**
 * Store Locator Application Logic (Commercial Template Setup)
 */

// CENTRAL CLIENT CONFIGURATION LAYER
// When onboarding a new client, modify only these values:
const CLIENT_CONFIG = {
    brandName: "Austin Location Finder",
    tagline: "Find the closest drop-off location instantly.",
    defaultCoords: { lat: 30.2672, lng: -97.7431 }, // Back to Austin coords
    mapZoom: 12,
    radiusMeters: 10000 
};

let map;
let infoWindow;
let markers = [];
const BACKEND_API_URL = '/api/search'; 

// CRITICAL FIX: Expose initMap globally for the script callback to register properly
window.initMap = async function() {
    console.log("Google Maps API loaded. Initializing client interface module...");
    
    // Inject brand identities dynamically straight into the viewport DOM elements
    document.title = CLIENT_CONFIG.brandName;
    const headerTitle = document.querySelector('.sidebar header h1');
    const headerTagline = document.querySelector('.sidebar header p');
    if (headerTitle) headerTitle.textContent = CLIENT_CONFIG.brandName;
    if (headerTagline) headerTagline.textContent = CLIENT_CONFIG.tagline;

    try {
        // Asynchronously request core map layouts and marker library assets
        const { Map } = await google.maps.importLibrary("maps");
        await google.maps.importLibrary("marker");

        map = new Map(document.getElementById('map'), {
            center: CLIENT_CONFIG.defaultCoords,
            zoom: CLIENT_CONFIG.mapZoom,
            mapId: 'DEMO_MAP_ID', // Replaced with evaluation token for AdvancedMarkerElement support
        });

        infoWindow = new google.maps.InfoWindow();
        
        const searchButton = document.getElementById('search-button');
        if (searchButton) {
            searchButton.addEventListener('click', handleSearch);
        } else {
            console.error("Critical Runtime Error: Search controller button missing.");
        }
        
        // Initial data rendering pull using the defaults
        fetchStores(CLIENT_CONFIG.defaultCoords);
    } catch (err) {
        console.error("Failed to initialize Google Maps interface components:", err);
        alertMessage("Failed to initialize Google Maps elements. Verify API configuration privileges.", "error");
    }
}

/**
 * Handles the user search query event block loop sequence
 */
function handleSearch() {
    const addressInput = document.getElementById('address-input');
    const address = addressInput.value.trim();

    if (!address) {
        alertMessage('Please enter a valid address string or location query.', 'warning');
        return;
    }

    const geocoder = new google.maps.Geocoder();
    geocoder.geocode({ address: address }, (results, status) => {
        if (status === 'OK' && results[0]) {
            const location = results[0].geometry.location;
            
            map.setCenter(location);
            map.setZoom(13);
            
            fetchStores({ lat: location.lat(), lng: location.lng() });
        } else {
            alertMessage(`Geocoding lookup execution path failed: "${address}". Status: ${status}`, 'error');
        }
    });
}

/**
 * Accesses local microservice endpoint layer using current spatial coordinates
 */
async function fetchStores(centerCoords) {
    const { lat, lng } = centerCoords;
    const url = `${BACKEND_API_URL}?lat=${lat}&lng=${lng}&radius=${CLIENT_CONFIG.radiusMeters}`; 

    const listElement = document.getElementById('store-list');
    listElement.innerHTML = '<p class="p-4 text-center text-gray-500">Searching active location profile registers...</p>';
    clearMarkers();
    
    try {
        const response = await fetch(url);
        if (!response.ok) throw new Error(`HTTP network runtime error! status: ${response.status}`);
        
        const geoJson = await response.json(); 
        
        if (geoJson.status === 'ok' && geoJson.features && geoJson.features.length > 0) {
             displayFeatures(geoJson.features);
             alertMessage(`${geoJson.features.length} unique profiles discovered!`, 'success');
        } else if (geoJson.status === 'ok' && geoJson.features.length === 0) {
             const boundaryRadiusKm = CLIENT_CONFIG.radiusMeters / 1000;
             listElement.innerHTML = `<p class="p-4 text-center text-gray-500">Zero points found within the baseline ${boundaryRadiusKm} km range parameters.</p>`;
        } else {
             throw new Error(geoJson.error || "Microservice backend failure.");
        }
    } catch (error) {
        console.error("Runtime data fetching exception logged:", error);
        alertMessage(`⚠️ System Interface Fault: ${error.message}. Ensure backend runtime microservice handles requests properly.`, 'error');
    }
}

function clearMarkers() {
    markers.forEach(marker => marker.map = null); 
    markers = [];
}

function displayFeatures(features) {
    const listElement = document.getElementById('store-list');
    listElement.innerHTML = ''; 

    features.forEach((feature, index) => {
        const properties = feature.properties;
        const coords = feature.geometry.coordinates;
        const position = { lat: coords[1], lng: coords[0] }; 

        // 1. Render modern Advanced Marker objects inside active Google Viewport
        const marker = new google.maps.marker.AdvancedMarkerElement({
            position: position,
            map: map,
            title: properties.business_name || properties.name,
            content: createMarkerContent(index + 1),
        });

        const storeData = {
            name: properties.business_name || properties.name,
            address: properties.address_address || properties.address || "No structural address string specified",
            distance: properties.distance_km,
            hours: properties.phone || "Operational (Mock Range)"
        };

        // CRITICAL UPDATE: Utilizing native 'gmp-click' for Web Component safety
        marker.addListener('gmp-click', () => {
            showStoreDetails(marker, storeData);
        });

        markers.push(marker);

        // 2. Append visual layout lists directly inside the side UI column element block
        const card = document.createElement('div');
        card.className = 'store-card';
        card.setAttribute('data-index', index);
        card.innerHTML = `
            <div class="store-info">
                <h3>${storeData.name}</h3>
                <p>${storeData.address}</p>
                <p class="text-xs text-blue-600">${storeData.hours}</p>
            </div>
            <div class="distance-info">
                <span class="font-bold text-lg">${storeData.distance}</span>
                <span class="text-sm text-gray-500">km</span>
            </div>
        `;
        
        card.addEventListener('click', () => {
            map.setCenter(position);
            map.setZoom(14);
            showStoreDetails(marker, storeData);
        });

        listElement.appendChild(card);
    });
}

function createMarkerContent(label) {
    const pin = document.createElement('div');
    pin.className = 'pin-label';
    pin.textContent = label;
    return pin;
}

function showStoreDetails(marker, store) {
    const content = `
        <div class="p-2">
            <h4 class="font-bold text-lg">${store.name}</h4>
            <p class="text-sm text-gray-700 mb-2">${store.address}</p>
            <hr class="my-2">
            <p class="text-xs"><strong>Contact Line:</strong> ${store.hours}</p>
            <p class="text-xs"><strong>Geospatial Range:</strong> ${store.distance} km away</p>
        </div>
    `;
    infoWindow.setContent(content);
    infoWindow.open(map, marker);
}

function alertMessage(message, type) {
    console.warn(`[${type.toUpperCase()}] ${message}`);
    const list = document.getElementById('store-list');
    const msgElement = document.createElement('p');
    msgElement.className = `p-4 font-semibold ${type === 'error' ? 'text-red-600' : (type === 'success' ? 'text-green-600' : 'text-yellow-600')}`;
    msgElement.textContent = message;
    
    list.prepend(msgElement);
    setTimeout(() => { msgElement.remove(); }, 5000);
}