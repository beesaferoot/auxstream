import { useState, useRef } from 'react'
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  Button,
  VStack,
  Text,
  Input,
  FormControl,
  FormLabel,
  FormErrorMessage,
  Box,
  Progress,
  HStack,
  Icon,
  useToast,
} from '@chakra-ui/react'
import { IoCloudUpload, IoMusicalNotes, IoCheckmarkCircle } from 'react-icons/io5'
import { uploadTrack, createArtist } from '../utils/api'

interface UploadTrackModalProps {
  isOpen: boolean
  onClose: () => void
  onUploadSuccess?: () => void
}

const UploadTrackModal = ({
  isOpen,
  onClose,
  onUploadSuccess,
}: UploadTrackModalProps) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [title, setTitle] = useState('')
  const [artist, setArtist] = useState('')
  const [isUploading, setIsUploading] = useState(false)
  const [uploadProgress, setUploadProgress] = useState(0)
  const [isDragging, setIsDragging] = useState(false)
  const [errors, setErrors] = useState<{ title?: string; artist?: string; file?: string }>({})
  const fileInputRef = useRef<HTMLInputElement>(null)
  const toast = useToast()

  const acceptedFormats = ['.mp3', '.wav', '.ogg', '.m4a', '.flac']
  const maxFileSize = 100 * 1024 * 1024 // 100MB

  const validateFile = (file: File): string | null => {
    const extension = '.' + file.name.split('.').pop()?.toLowerCase()
    if (!acceptedFormats.includes(extension)) {
      return `Please upload a valid audio file (${acceptedFormats.join(', ')})`
    }
    if (file.size > maxFileSize) {
      return 'File size must be less than 100MB'
    }
    return null
  }

  const handleFileSelect = (file: File) => {
    const error = validateFile(file)
    if (error) {
      setErrors({ ...errors, file: error })
      toast({
        title: 'Invalid file',
        description: error,
        status: 'error',
        duration: 4000,
        isClosable: true,
      })
      return
    }

    setSelectedFile(file)
    setErrors({ ...errors, file: undefined })

    // Auto-populate title from filename if empty
    if (!title) {
      const filename = file.name.replace(/\.[^/.]+$/, '') // Remove extension
      setTitle(filename)
    }
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)

    const files = e.dataTransfer.files
    if (files.length > 0) {
      handleFileSelect(files[0])
    }
  }

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (files && files.length > 0) {
      handleFileSelect(files[0])
    }
  }

  const validateForm = (): boolean => {
    const newErrors: { title?: string; artist?: string; file?: string } = {}

    if (!title.trim()) {
      newErrors.title = 'Title is required'
    }
    if (!artist.trim()) {
      newErrors.artist = 'Artist name is required'
    }
    if (!selectedFile) {
      newErrors.file = 'Please select an audio file'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleUpload = async () => {
    if (!validateForm() || !selectedFile) return

    setIsUploading(true)
    setUploadProgress(0)

    try {
      // First, create/get the artist
      const artistData = await createArtist(artist.trim())
      
      // Then upload the track with the artist ID
      await uploadTrack(
        selectedFile,
        { 
          title: title.trim(), 
          artistId: artistData.id,
        },
        (progress) => {
          setUploadProgress(progress)
        }
      )

      toast({
        title: 'Upload successful!',
        description: `"${title}" has been uploaded successfully.`,
        status: 'success',
        duration: 5000,
        isClosable: true,
      })

      // Reset form
      setSelectedFile(null)
      setTitle('')
      setArtist('')
      setUploadProgress(0)

      if (onUploadSuccess) {
        onUploadSuccess()
      }

      onClose()
    } catch (error) {
      toast({
        title: 'Upload failed',
        description: error instanceof Error ? error.message : 'An error occurred',
        status: 'error',
        duration: 5000,
        isClosable: true,
      })
    } finally {
      setIsUploading(false)
    }
  }

  const handleClose = () => {
    if (!isUploading) {
      setSelectedFile(null)
      setTitle('')
      setArtist('')
      setUploadProgress(0)
      setErrors({})
      onClose()
    }
  }

  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
  }

  return (
    <Modal isOpen={isOpen} onClose={handleClose} size="xl" closeOnOverlayClick={!isUploading}>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Upload Audio Track</ModalHeader>
        <ModalCloseButton isDisabled={isUploading} />
        <ModalBody>
          <VStack spacing={6} align="stretch">
            {/* File Drop Zone */}
            <Box>
              <FormControl isInvalid={!!errors.file}>
                <FormLabel>Audio File</FormLabel>
                <Box
                  border="2px dashed"
                  borderColor={isDragging ? 'brand.500' : errors.file ? 'red.500' : 'gray.300'}
                  borderRadius="lg"
                  p={8}
                  textAlign="center"
                  bg={isDragging ? 'blue.50' : 'gray.50'}
                  cursor="pointer"
                  transition="all 0.2s"
                  _hover={{ borderColor: 'brand.500', bg: 'blue.50' }}
                  onDragOver={handleDragOver}
                  onDragLeave={handleDragLeave}
                  onDrop={handleDrop}
                  onClick={() => fileInputRef.current?.click()}
                >
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept={acceptedFormats.join(',')}
                    onChange={handleFileInputChange}
                    style={{ display: 'none' }}
                    disabled={isUploading}
                  />

                  {selectedFile ? (
                    <VStack spacing={2}>
                      <Icon as={IoCheckmarkCircle} boxSize={12} color="green.500" />
                      <Text fontWeight="bold" fontSize="lg">
                        {selectedFile.name}
                      </Text>
                      <Text fontSize="sm" color="gray.600">
                        {formatFileSize(selectedFile.size)}
                      </Text>
                      {!isUploading && (
                        <Button size="sm" variant="ghost" colorScheme="blue">
                          Change File
                        </Button>
                      )}
                    </VStack>
                  ) : (
                    <VStack spacing={2}>
                      <Icon as={IoCloudUpload} boxSize={12} color="gray.400" />
                      <Text fontWeight="medium">
                        Drag and drop your audio file here
                      </Text>
                      <Text fontSize="sm" color="gray.600">
                        or click to browse
                      </Text>
                      <Text fontSize="xs" color="gray.500">
                        Supported formats: {acceptedFormats.join(', ')} (max 100MB)
                      </Text>
                    </VStack>
                  )}
                </Box>
                {errors.file && <FormErrorMessage>{errors.file}</FormErrorMessage>}
              </FormControl>
            </Box>

            {/* Track Metadata */}
            <FormControl isInvalid={!!errors.title} isRequired>
              <FormLabel>Track Title</FormLabel>
              <Input
                placeholder="Enter track title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                isDisabled={isUploading}
              />
              {errors.title && <FormErrorMessage>{errors.title}</FormErrorMessage>}
            </FormControl>

            <FormControl isInvalid={!!errors.artist} isRequired>
              <FormLabel>Artist Name</FormLabel>
              <Input
                placeholder="Enter artist name"
                value={artist}
                onChange={(e) => setArtist(e.target.value)}
                isDisabled={isUploading}
              />
              {errors.artist && <FormErrorMessage>{errors.artist}</FormErrorMessage>}
            </FormControl>

            {/* Upload Progress */}
            {isUploading && (
              <Box>
                <HStack justify="space-between" mb={2}>
                  <Text fontSize="sm" fontWeight="medium">
                    Uploading...
                  </Text>
                  <Text fontSize="sm" color="gray.600">
                    {Math.round(uploadProgress)}%
                  </Text>
                </HStack>
                <Progress
                  value={uploadProgress}
                  size="sm"
                  colorScheme="brand"
                  borderRadius="full"
                  hasStripe
                  isAnimated
                />
              </Box>
            )}

            {/* Preview Box */}
            {selectedFile && !isUploading && (
              <Box
                bg="gray.50"
                p={4}
                borderRadius="md"
                border="1px"
                borderColor="gray.200"
              >
                <HStack spacing={3}>
                  <Icon as={IoMusicalNotes} boxSize={6} color="brand.500" />
                  <VStack align="start" spacing={0} flex={1}>
                    <Text fontWeight="bold" fontSize="sm">
                      {title || 'Untitled'}
                    </Text>
                    <Text fontSize="xs" color="gray.600">
                      {artist || 'Unknown Artist'}
                    </Text>
                  </VStack>
                </HStack>
              </Box>
            )}
          </VStack>
        </ModalBody>

        <ModalFooter>
          <Button variant="ghost" mr={3} onClick={handleClose} isDisabled={isUploading}>
            Cancel
          </Button>
          <Button
            colorScheme="brand"
            onClick={handleUpload}
            isLoading={isUploading}
            loadingText="Uploading..."
            isDisabled={!selectedFile}
            leftIcon={<IoCloudUpload />}
          >
            Upload Track
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}

export default UploadTrackModal

