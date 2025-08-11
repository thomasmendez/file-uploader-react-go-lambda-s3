import { useState, useRef } from 'react'
import './App.css'

function App() {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploadStatus, setUploadStatus] = useState<string>('');

  const handleButtonClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files && event.target.files[0];
    if (!file) {
      return;
    }
    setSelectedFile(file);
    setUploadStatus('');
  };

  // Method 1: Raw Binary Data
  const uploadRawBinary = async () => {
    if (!selectedFile) {
      setUploadStatus('Please select a file first');
      return;
    }

    try {
      setUploadStatus('Uploading raw binary...');
      
      const response = await fetch('http://127.0.0.1:3000/api/upload/binary', {
        method: 'POST',
        headers: {
          'Content-Type': selectedFile.type,
          'X-Filename': selectedFile.name, // Send filename in header
        },
        body: selectedFile // Send file directly as binary data
      });

      if (!response.ok) {
        const errorText = await response.text();
        console.error('Raw binary upload error:', errorText);
        throw new Error(`Upload failed: ${response.statusText}`);
      }
      
      const result = await response.json();
      console.log('Raw binary upload successful:', result);
      setUploadStatus('Raw binary upload successful!');
    } catch (error) {
      console.error('Raw binary upload error:', error);
      setUploadStatus('Raw binary upload failed');
    }
  };

  // Method 2: Multipart Form Data
  const uploadFormData = async () => {
    if (!selectedFile) {
      setUploadStatus('Please select a file first');
      return;
    }

    try {
      setUploadStatus('Uploading form data...');
      
      const formData = new FormData();
      formData.append('image', selectedFile);
      formData.append('filename', selectedFile.name);

      const response = await fetch('http://127.0.0.1:3000/api/upload/formdata', {
        method: 'POST',
        // Don't set Content-Type header - let browser set it with boundary
        body: formData
      });

      if (!response.ok) {
        const errorText = await response.text();
        console.error('Form data upload error:', errorText);
        throw new Error(`Upload failed: ${response.statusText}`);
      }
      
      const result = await response.json();
      console.log('Form data upload successful:', result);
      setUploadStatus('Form data upload successful!');
    } catch (error) {
      console.error('Form data upload error:', error);
      setUploadStatus('Form data upload failed');
    }
  };

  // Method 3: Base64 Encoded
  const uploadBase64 = async () => {
    if (!selectedFile) {
      setUploadStatus('Please select a file first');
      return;
    }

    try {
      setUploadStatus('Uploading base64...');
      
      const fileReader = new FileReader();
      fileReader.onload = async () => {
        try {
          const base64String = fileReader.result as string;
          const base64Data = base64String.split(',')[1]; // Remove data:image/jpeg;base64, prefix

          console.log('File type:', selectedFile.type);
          console.log('File size:', selectedFile.size);
          console.log('Base64 length:', base64Data.length);

          const response = await fetch('http://127.0.0.1:3000/api/upload/base64', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              imageData: base64Data,
              contentType: selectedFile.type,
              filename: selectedFile.name
            }),
          });

          if (!response.ok) {
            const errorText = await response.text();
            console.error('Base64 upload error:', errorText);
            throw new Error(`Upload failed: ${response.statusText}`);
          }
          
          const result = await response.json();
          console.log('Base64 upload successful:', result);
          setUploadStatus('Base64 upload successful!');
        } catch (error) {
          console.error('Base64 upload error:', error);
          setUploadStatus('Base64 upload failed');
        }
      };

      fileReader.readAsDataURL(selectedFile);
    } catch (error) {
      console.error('Base64 upload error:', error);
      setUploadStatus('Base64 upload failed');
    }
  };

  return (
    <>
      <div style={{ padding: '20px', fontFamily: 'Arial, sans-serif' }}>
        <h1>Image Upload Methods Test</h1>
        
        {/* File Selection */}
        <div style={{ marginBottom: '30px', padding: '20px', border: '2px solid #ccc', borderRadius: '8px' }}>
          <h2>File Selection</h2>
          <input
            type="file"
            ref={fileInputRef}
            onChange={handleFileChange}
            accept="image/*"
            style={{ display: 'none' }}
          />
          <button 
            onClick={handleButtonClick}
            style={{ padding: '10px 20px', fontSize: '16px', cursor: 'pointer' }}
          >
            Select Image File
          </button>
          {selectedFile && (
            <p style={{ marginTop: '10px', color: '#333' }}>
              Selected: {selectedFile.name} ({(selectedFile.size / 1024).toFixed(2)} KB)
            </p>
          )}
        </div>

        {/* Upload Methods - Horizontal Layout */}
        <div style={{ display: 'flex', gap: '20px', marginBottom: '20px', flexWrap: 'wrap' }}>
          {/* Method 1: Raw Binary */}
          <div style={{ flex: '1', minWidth: '280px', padding: '20px', border: '2px solid #ff6b6b', borderRadius: '8px' }}>
            <h2 style={{ color: '#ff6b6b', margin: '0 0 10px 0', fontSize: '18px' }}>Method 1: Raw Binary Data</h2>
            <p style={{ margin: '0 0 15px 0', fontSize: '14px', lineHeight: '1.4' }}>Sends the file directly as binary data in the request body</p>
            <button 
              onClick={uploadRawBinary}
              disabled={!selectedFile}
              style={{ 
                padding: '10px 20px', 
                fontSize: '14px', 
                cursor: selectedFile ? 'pointer' : 'not-allowed',
                backgroundColor: '#ff6b6b',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                width: '100%'
              }}
            >
              Upload Raw Binary
            </button>
          </div>

          {/* Method 2: Form Data */}
          <div style={{ flex: '1', minWidth: '280px', padding: '20px', border: '2px solid #4ecdc4', borderRadius: '8px' }}>
            <h2 style={{ color: '#4ecdc4', margin: '0 0 10px 0', fontSize: '18px' }}>Method 2: Multipart Form Data</h2>
            <p style={{ margin: '0 0 15px 0', fontSize: '14px', lineHeight: '1.4' }}>Uses FormData for multipart/form-data encoding (standard file upload)</p>
            <button 
              onClick={uploadFormData}
              disabled={!selectedFile}
              style={{ 
                padding: '10px 20px', 
                fontSize: '14px', 
                cursor: selectedFile ? 'pointer' : 'not-allowed',
                backgroundColor: '#4ecdc4',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                width: '100%'
              }}
            >
              Upload Form Data
            </button>
          </div>

          {/* Method 3: Base64 */}
          <div style={{ flex: '1', minWidth: '280px', padding: '20px', border: '2px solid #45b7d1', borderRadius: '8px' }}>
            <h2 style={{ color: '#45b7d1', margin: '0 0 10px 0', fontSize: '18px' }}>Method 3: Base64 Encoded</h2>
            <p style={{ margin: '0 0 15px 0', fontSize: '14px', lineHeight: '1.4' }}>Converts image to base64 string and sends as JSON</p>
            <button 
              onClick={uploadBase64}
              disabled={!selectedFile}
              style={{ 
                padding: '10px 20px', 
                fontSize: '14px', 
                cursor: selectedFile ? 'pointer' : 'not-allowed',
                backgroundColor: '#45b7d1',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                width: '100%'
              }}
            >
              Upload Base64
            </button>
          </div>
        </div>

        {/* Status Display */}
        {uploadStatus && (
          <div style={{ 
            marginTop: '20px', 
            padding: '15px', 
            backgroundColor: uploadStatus.includes('successful') ? '#d4edda' : '#f8d7da',
            color: uploadStatus.includes('successful') ? '#155724' : '#721c24',
            border: `1px solid ${uploadStatus.includes('successful') ? '#c3e6cb' : '#f5c6cb'}`,
            borderRadius: '4px'
          }}>
            <strong>Status:</strong> {uploadStatus}
          </div>
        )}
      </div>
    </>
  )
}

export default App