window.onload = () => {
    const enableCameraCheckbox = document.querySelector('.barcode-scanner .ctrl.enable');
    const frontCameraCheckbox = document.querySelector('.barcode-scanner .ctrl.front');
    const zoomCameraRange = document.querySelector('.barcode-scanner .ctrl.zoom');
    const cameraVideo = document.querySelector('.barcode-scanner video');
    const scannerLogSpan = document.querySelector('.barcode-scanner .log');
    const scannerIdInput = document.querySelector('.barcode-scanner .scanner-id');
    const boardIdInput = document.querySelector('#board-id');

    const log = (text) => {
        scannerLogSpan.innerText = text;
    };
    const logError = (message) => (error) => {
        log(message + ': ' + error);
    }
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
    const handleBarcodes = (barcodes) => {
        if (barcodes.length == 1) {
            const barcode = barcodes[0].rawValue;
            boardIdInput.value = barcode;
            log('scanned board id: ' + barcode);
        }
    };
    const handleImageBitmap = (barcodeDetector) => (imageBitmap) => {
        barcodeDetector?.detect(imageBitmap)
            .then(handleBarcodes)
            .catch(logError('detecting bar codes'));
    };
    const scanBarcode = (imageCapture, barcodeDetector) => () => {
        if (!imageCapture.track.muted) {
            imageCapture.grabFrame()
                .then(handleImageBitmap(barcodeDetector));
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
    const handleMediaStream = (barcodeDetector) => (mediaStream) => {
        cameraVideo.srcObject = mediaStream;
        const track = getTrack();
        const imageCapture = new ImageCapture(track);
        scannerIdInput.value = setInterval(scanBarcode(imageCapture, barcodeDetector), 250);
        frontCameraCheckbox.hidden = false;
        initCameraZoom(track);
        cameraVideo.hidden = false;
        log('starting video capture');
    }
    const startVideo = (barcodeDetector) => {
        stopVideo();
        const facingMode = frontCameraCheckbox.checked ? 'user' : 'environment';
        const constraints = { video: { facingMode } };
        navigator.mediaDevices.getUserMedia(constraints)
            .then(handleMediaStream(barcodeDetector))
            .catch(logError('camera not found'));
    };
    const handleSupportedBarcodeFormats = (supportedFormats) => {
        const formats = supportedFormats.filter(format => ['qr_code', 'aztec', 'data_matrix'].includes(format));
        if (formats.length == 0) {
            log('browser cannot detect any type of bar code on board');
        } else {
            log('browser can detect ' + formats.join(', ') + ' bar code types');
            const barcodeDetector = new BarcodeDetector({ formats });
            enableCameraCheckbox.onclick = () => { enableCameraCheckbox.checked ? startVideo(barcodeDetector) : stopVideo() };
            frontCameraCheckbox.onclick = enableCameraCheckbox.onclick;
            zoomCameraRange.oninput = (event) => getTrack().applyConstraints({ advanced: [{ zoom: event.target.value }] });
            enableCameraCheckbox.hidden = false;
        }
    }
    const init = () => {
        if ('BarcodeDetector' in window) {
            BarcodeDetector.getSupportedFormats()
                .then(handleSupportedBarcodeFormats);
        } else {
            log('browser cannot detect bar code on board');
        }
    };
    init();
};