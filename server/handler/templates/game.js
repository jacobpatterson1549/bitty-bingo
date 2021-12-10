window.onload = () => {
    const enableCameraCheckbox = document.querySelector('.bar-code-scanner .ctrl.enable');
    const frontCameraCheckbox = document.querySelector('.bar-code-scanner .ctrl.front');
    const zoomCameraRange = document.querySelector('.bar-code-scanner .ctrl.zoom');
    const cameraVideo = document.querySelector('.bar-code-scanner video');
    const scannerLogSpan = document.querySelector('.bar-code-scanner .log');
    const scannerIdInput = document.querySelector('.bar-code-scanner .scanner-id');
    const boardIdInput = document.querySelector('#board-id');

    const log = (text) => {
        scannerLogSpan.innerText = text;
    };
    const initCameraZoom = (track) => {
        const trackSettings = track.getSettings();
        if ('zoom' in trackSettings) {
            const trackCapabilities = track.getCapabilities();
            zoomCameraRange.min = trackCapabilities.zoom.min;
            zoomCameraRange.max = trackCapabilities.zoom.max;
            zoomCameraRange.step = trackCapabilities.zoom.step;
            zoomCameraRange.value = trackSettings.zoom;
            zoomCameraRange.hidden = false;
        } else {
            log('camera zoom not supported');
        }
    };
    const getTrack = () => {
        return cameraVideo.srcObject?.getVideoTracks()[0];
    };
    const scanQR = (imageCapture, barcodeDetector) => () => {
        if (!imageCapture.track.muted) {
            imageCapture.grabFrame()
                .then(imageBitmap => {
                    barcodeDetector?.detect(imageBitmap)
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
        }
    };
    const stopVideo = () => {
        log('stopping video capture');
        getTrack()?.stop(); // turn off camera
        cameraVideo.srcObject = null;
        clearInterval(scannerIdInput.value);
        frontCameraCheckbox.hidden = true;
        zoomCameraRange.hidden = true;
        cameraVideo.hidden = true;
    };
    const startVideo = (barcodeDetector) => {
        stopVideo();
        const facingMode = frontCameraCheckbox.checked ? 'user' : 'environment';
        const constraints = { video: { facingMode } };
        navigator.mediaDevices.getUserMedia(constraints)
            .then(mediaStream => {
                cameraVideo.srcObject = mediaStream;
                const track = getTrack();
                const imageCapture = new ImageCapture(track);
                scannerIdInput.value = setInterval(scanQR(imageCapture, barcodeDetector), 250);
                frontCameraCheckbox.hidden = false;
                initCameraZoom(track);
                cameraVideo.hidden = false;
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
                        const barcodeDetector = new BarcodeDetector({ formats });
                        enableCameraCheckbox.onclick = () => { enableCameraCheckbox.checked ? startVideo(barcodeDetector) : stopVideo() };
                        frontCameraCheckbox.onclick = enableCameraCheckbox.onclick;
                        zoomCameraRange.oninput = (event) => getTrack().applyConstraints({ advanced: [{ zoom: event.target.value }] });
                        enableCameraCheckbox.hidden = false;
                    }
                });
        } else {
            log('browser cannot detect bar code on board');
        }
    };
    init();
};