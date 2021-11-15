const checkBoard = {
    videoEnabled: false,
    track: null,
    imageCapture: null,
    scannerID: null,
    barcodeDetector: null,
    frontCamera: true,
    toggleVideo: ()  => {
        checkBoard.videoEnabled ? checkBoard.stopVideo() : checkBoard.startVideo();
        checkBoard.videoEnabled = !checkBoard.videoEnabled;
    },
    flipCamera: () => {
        checkBoard.frontCamera = !checkBoard.frontCamera;
        if (checkBoard.videoEnabled) {
            checkBoard.startVideo();
        }
    },
    startVideo: () => {
        checkBoard.stopVideo();
        checkBoard.log('starting video capture');
        const toggleVideoButton = document.querySelector('#check-board .toggle-video');
        const flipCameraButton = document.querySelector('#check-board .flip-camera');
        toggleVideoButton.disabled = true;
        toggleVideoButton.value = toggleVideoButton.value.replace('Enable', 'Disable');
        const facingMode = checkBoard.frontCamera ? "environment": "user";
        const constraints = { video: { facingMode } };
        navigator.mediaDevices.getUserMedia(constraints)
            .then(mediaStream => {
                const video = document.querySelector('#check-board video');
                video.srcObject = mediaStream;
                checkBoard.track = mediaStream.getVideoTracks()[0];
                checkBoard.imageCapture = new ImageCapture(checkBoard.track);
                checkBoard.scannerID = setInterval(checkBoard.scanQR, 250);
                checkBoard.initCameraZoom();
                video.hidden = false;
                toggleVideoButton.disabled = false;
                flipCameraButton.hidden = false;
            })
            .catch(error => {
                checkBoard.log('camera not found: ' + error);
            });
    },
    stopVideo: () => {
        checkBoard.log('stopping video capture');
        const video = document.querySelector('#check-board video');
        const toggleVideoButton = document.querySelector('#check-board .toggle-video');
        const flipCameraButton = document.querySelector('#check-board .flip-camera');
        const cameraZoomRange = document.querySelector('#check-board .camera-zoom');
        
        video.srcObject = null;
        toggleVideoButton.value = toggleVideoButton.value.replace('Disable', 'Enable');
        checkBoard.track?.stop(); // turn off camera
        clearInterval(checkBoard.scannerID);
        video.hidden = true;
        flipCameraButton.hidden = true;
        cameraZoomRange.hidden = true;
    },
    scanQR: () => {
        checkBoard.imageCapture.grabFrame()
            .then(imageBitmap => {
                checkBoard.barcodeDetector.detect(imageBitmap)
                    .then(barCodes => {
                        if (barCodes.length != 1) {
                            return;
                        }
                        const qrCode = barCodes[0].rawValue;
                        const boardIDInput = document.querySelector('#check-board input[name="boardID"]');
                        boardIDInput.value = qrCode;
                        checkBoard.log('scanned board id: ' + qrCode);
                    })
                    .catch(error => {
                        checkBoard.log('detecting bar codes: ' + error);
                    })
            });
    },
    initCameraZoom: () => {
        const trackSettings = checkBoard?.track.getSettings();
        if (!('zoom' in trackSettings)) {
            checkBoard.log('camera zoom not supported');
            return;
        }
        const trackCapabilities = checkBoard?.track.getCapabilities();
        const cameraZoomRange = document.querySelector('#check-board .camera-zoom');
        cameraZoomRange.min = trackCapabilities.zoom.min;
        cameraZoomRange.max = trackCapabilities.zoom.max;
        cameraZoomRange.step = trackCapabilities.zoom.step;
        cameraZoomRange.value = trackSettings.zoom;
        cameraZoomRange.hidden = false;
    },
    setCameraZoom: (event) => {
        const zoom = { advanced: [ { zoom: event.target.value } ] };
        checkBoard?.track.applyConstraints(zoom);
    },
    log: (text) => {
        const span = document.querySelector('#check-board .log');
        span.innerText = text;
    },
    init: () => {
        if (!('BarcodeDetector' in window)) {
            checkBoard.log('browser cannot detect bar code on board');
            return;
        }
        BarcodeDetector.getSupportedFormats()
            .then(supportedFormats => {
                const formats = supportedFormats.filter(format => format === 'qr_code');
                if (formats.length == 0) {
                    checkBoard.log('browser cannot detect QR code on board');
                    return;
                }
                checkBoard.barcodeDetector = new BarcodeDetector({formats});
                const toggleVideoButton = document.querySelector('#check-board .toggle-video');
                const flipCameraButton = document.querySelector('#check-board .flip-camera');
                const cameraZoomRange = document.querySelector('#check-board .camera-zoom');
                const checkbox = document.querySelector('#check-board .hide-qr-controls');
                toggleVideoButton.onclick = checkBoard.toggleVideo;
                flipCameraButton.onclick = checkBoard.flipCamera
                cameraZoomRange.oninput = checkBoard.setCameraZoom;
                checkbox.checked = false;
            });
    },
};
checkBoard.init();