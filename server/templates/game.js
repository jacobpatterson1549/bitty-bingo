(() => {
    const allowCameraCheckbox = document.querySelector('#allow-camera-checkbox');
    const enableCameraCheckbox = document.querySelector('#enable-camera-checkbox');
    const frontCameraCheckbox = document.querySelector('#front-camera-checkbox');
    const cameraZoomRange = document.querySelector('#camera-zoom-range');
    const cameraLog = document.querySelector('#camera-log');
    const video = document.querySelector('#check-board video');
    const boardIdInput = document.querySelector('#check-board input[name="boardID"]');

    let track = null;
    let imageCapture = null;
    let scannerID = null;
    let barcodeDetector = null;

    const log = (text) => {
        cameraLog.innerText = text;
    };
    const initCameraZoom = () => {
        const trackSettings = track.getSettings();
        if ('zoom' in trackSettings) {
            const trackCapabilities = track.getCapabilities();
            cameraZoomRange.min = trackCapabilities.zoom.min;
            cameraZoomRange.max = trackCapabilities.zoom.max;
            cameraZoomRange.step = trackCapabilities.zoom.step;
            cameraZoomRange.value = trackSettings.zoom;
            cameraZoomRange.hidden = false;
        } else {
            log('camera zoom not supported');
        }
    };
    const scanQR = () => {
        imageCapture.grabFrame()
            .then(imageBitmap => {
                barcodeDetector.detect(imageBitmap)
                    .then(barCodes => {
                        if (barCodes.length == 1) {
                            const qrCode = barCodes[0].rawValue;
                            boardIdInput.value = qrCode;
                            log('scanned board id: ' + qrCode);
                        }
                    })
                    .catch(error => {
                        log('detecting bar codes: ' + error);
                    });
            });
    };
    const stopVideo = () => {
        log('stopping video capture');
        video.srcObject = null;
        track?.stop(); // turn off camera
        clearInterval(scannerID);
        frontCameraCheckbox.hidden = true;
        cameraZoomRange.hidden = true;
        video.hidden = true;
    };
    const startVideo = () => {
        stopVideo();
        const facingMode = frontCameraCheckbox.checked ? "user" : "environment";
        const constraints = { video: { facingMode } };
        navigator.mediaDevices.getUserMedia(constraints)
            .then(mediaStream => {
                video.srcObject = mediaStream;
                track = mediaStream.getVideoTracks()[0];
                imageCapture = new ImageCapture(track);
                scannerID = setInterval(scanQR, 250);
                frontCameraCheckbox.hidden = false;
                initCameraZoom();
                video.hidden = false;
                log('starting video capture');
            })
            .catch(error => {
                log('camera not found: ' + error);
            });
    };
    const init = () => {
        if ('BarcodeDetector' in window) {
            BarcodeDetector.getSupportedFormats()
                .then(supportedFormats => {
                    const formats = supportedFormats.filter(format => format === 'qr_code');
                    if (formats.length == 0) {
                        log('browser cannot detect QR code on board');
                    } else {
                        barcodeDetector = new BarcodeDetector({ formats });
                        enableCameraCheckbox.onclick = () => enableCameraCheckbox.checked ? startVideo() : stopVideo();
                        frontCameraCheckbox.onclick = () => enableCameraCheckbox.checked && startVideo();
                        cameraZoomRange.oninput = (event) => track.applyConstraints({ advanced: [{ zoom: event.target.value }] });
                        allowCameraCheckbox.checked = true;
                    }
                });
        } else {
            log('browser cannot detect bar code on board');
        }
    };
    init();
})();