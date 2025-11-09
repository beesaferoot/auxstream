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
  Table,
  Tbody,
  Tr,
  Td,
  IconButton,
  Badge,
} from '@chakra-ui/react'
import {
  IoCloudUpload,
  IoCheckmarkCircle,
  IoClose,
  IoAlert,
} from 'react-icons/io5'
import { uploadTracksBulk, createArtist } from '../utils/api'

interface UploadTrackModalProps {
  isOpen: boolean
  onClose: () => void
  onUploadSuccess?: () => void
}

interface FileWithMetadata {
  file: File
  title: string
  artist: string
  status: 'pending' | 'uploading' | 'success' | 'error'
  progress: number
  error?: string
}

const BulkUploadModal = ({
  isOpen,
  onClose,
  onUploadSuccess,
}: UploadTrackModalProps) => {
  const [files, setFiles] = useState<FileWithMetadata[]>([])
  const [defaultArtist, setDefaultArtist] = useState('')
  const [isDragging, setIsDragging] = useState(false)
  const [isUploading, setIsUploading] = useState(false)
  const [errors, setErrors] = useState<{ artist?: string; files?: string }>({})
  const fileInputRef = useRef<HTMLInputElement>(null)
  const toast = useToast()

  const acceptedFormats = ['.mp3', '.wav', '.ogg', '.m4a', '.flac']
  const maxFileSize = 100 * 1024 * 1024 // 100MB

  const validateFile = (file: File): string | null => {
    const extension = '.' + file.name.split('.').pop()?.toLowerCase()
    if (!acceptedFormats.includes(extension)) {
      return `Invalid format: ${extension}`
    }
    if (file.size > maxFileSize) {
      return 'File too large (max 100MB)'
    }
    return null
  }

  const handleFileSelect = (newFiles: FileList | null) => {
    if (!newFiles) return

    const validFiles: FileWithMetadata[] = []

    Array.from(newFiles).forEach((file) => {
      const error = validateFile(file)
      if (error) {
        toast({
          title: `Skipped: ${file.name}`,
          description: error,
          status: 'warning',
          duration: 3000,
          isClosable: true,
        })
        return
      }

      const filename = file.name.replace(/\.[^/.]+$/, '') // Remove extension
      validFiles.push({
        file,
        title: filename,
        artist: defaultArtist,
        status: 'pending',
        progress: 0,
      })
    })

    setFiles((prev) => [...prev, ...validFiles])
    setErrors({ ...errors, files: undefined })
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
    handleFileSelect(e.dataTransfer.files)
  }

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    handleFileSelect(e.target.files)
  }

  const removeFile = (index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index))
  }

  const updateFileMetadata = (
    index: number,
    field: 'title' | 'artist',
    value: string
  ) => {
    setFiles((prev) =>
      prev.map((f, i) => (i === index ? { ...f, [field]: value } : f))
    )
  }

  const validateForm = (): boolean => {
    const newErrors: { artist?: string; files?: string } = {}

    if (!defaultArtist.trim()) {
      newErrors.artist = 'Artist name is required'
    }

    if (files.length === 0) {
      newErrors.files = 'Please select at least one audio file'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const [overallProgress, setOverallProgress] = useState(0)

  const handleUploadAll = async () => {
    if (!validateForm()) return

    setIsUploading(true)
    setOverallProgress(0)

    // Mark all files as uploading
    setFiles((prev) =>
      prev.map((f) => ({
        ...f,
        artist: f.artist || defaultArtist,
        status: 'uploading',
        progress: 0,
      }))
    )

    try {
      // Create artist
      const artist = await createArtist(defaultArtist)

      // Prepare data for bulk upload
      const filesToUpload = files.map((f) => f.file)
      const titlesToUpload = files.map((f) => f.title)

      // Upload all files in one request
      const result = await uploadTracksBulk(
        filesToUpload,
        titlesToUpload,
        artist.id,
        (progress) => {
          setOverallProgress(progress)
          // Update all files with same progress
          setFiles((prev) =>
            prev.map((f) => ({ ...f, progress }))
          )
        }
      )

      // Mark all as success
      setFiles((prev) =>
        prev.map((f) => ({
          ...f,
          status: 'success',
          progress: 100,
        }))
      )

      toast({
        title: 'Bulk upload complete!',
        description: `Successfully uploaded ${result.data.rows} track(s)`,
        status: 'success',
        duration: 5000,
        isClosable: true,
      })

      if (onUploadSuccess) {
        onUploadSuccess()
      }

      // Auto-close after success
      setTimeout(() => {
        handleClose()
      }, 2000)
    } catch (error) {
      // Mark all as error
      const errorMsg =
        error instanceof Error ? error.message : 'Upload failed'

      setFiles((prev) =>
        prev.map((f) => ({
          ...f,
          status: 'error',
          error: errorMsg,
        }))
      )

      toast({
        title: 'Bulk upload failed',
        description: errorMsg,
        status: 'error',
        duration: 5000,
        isClosable: true,
      })
    } finally {
      setIsUploading(false)
      setOverallProgress(0)
    }
  }

  const handleClose = () => {
    if (!isUploading) {
      setFiles([])
      setDefaultArtist('')
      setErrors({})
      setOverallProgress(0)
      onClose()
    }
  }

  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
  }

  const getStatusIcon = (status: FileWithMetadata['status']) => {
    switch (status) {
      case 'success':
        return <IoCheckmarkCircle color="green" />
      case 'error':
        return <IoAlert color="red" />
      case 'uploading':
        return null
      default:
        return null
    }
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      size="2xl"
      closeOnOverlayClick={!isUploading}
    >
      <ModalOverlay />
      <ModalContent maxH="90vh">
        <ModalHeader>Bulk Upload Audio Tracks</ModalHeader>
        <ModalCloseButton isDisabled={isUploading} />
        <ModalBody overflowY="auto">
          <VStack spacing={6} align="stretch">
            {/* Default Artist */}
            <FormControl isInvalid={!!errors.artist} isRequired>
              <FormLabel>Artist Name (for all tracks)</FormLabel>
              <Input
                placeholder="Enter artist name"
                value={defaultArtist}
                onChange={(e) => {
                  setDefaultArtist(e.target.value)
                  // Update all files with new default artist
                  setFiles((prev) =>
                    prev.map((f) => ({ ...f, artist: e.target.value }))
                  )
                }}
                isDisabled={isUploading}
              />
              {errors.artist && (
                <FormErrorMessage>{errors.artist}</FormErrorMessage>
              )}
            </FormControl>

            {/* File Drop Zone */}
            <FormControl isInvalid={!!errors.files}>
              <FormLabel>Audio Files</FormLabel>
              <Box
                border="2px dashed"
                borderColor={
                  isDragging
                    ? 'brand.500'
                    : errors.files
                      ? 'red.500'
                      : 'gray.300'
                }
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
                  multiple
                  disabled={isUploading}
                />

                <VStack spacing={2}>
                  <Icon as={IoCloudUpload} boxSize={12} color="gray.400" />
                  <Text fontWeight="medium">
                    Drag and drop audio files here
                  </Text>
                  <Text fontSize="sm" color="gray.600">
                    or click to browse
                  </Text>
                  <Text fontSize="xs" color="gray.500">
                    Supported: {acceptedFormats.join(', ')} (max 100MB each)
                  </Text>
                </VStack>
              </Box>
              {errors.files && (
                <FormErrorMessage>{errors.files}</FormErrorMessage>
              )}
            </FormControl>

            {/* Overall Progress */}
            {isUploading && overallProgress > 0 && (
              <Box>
                <HStack justify="space-between" mb={2}>
                  <Text fontSize="sm" fontWeight="medium">
                    Uploading all files...
                  </Text>
                  <Text fontSize="sm" color="gray.600">
                    {Math.round(overallProgress)}%
                  </Text>
                </HStack>
                <Progress
                  value={overallProgress}
                  size="md"
                  colorScheme="brand"
                  borderRadius="full"
                  hasStripe
                  isAnimated
                />
              </Box>
            )}

            {/* File List */}
            {files.length > 0 && (
              <Box>
                <HStack justify="space-between" mb={2}>
                  <Text fontWeight="bold">{files.length} file(s) selected</Text>
                  <Text fontSize="sm" color="gray.600">
                    Total:{' '}
                    {formatFileSize(
                      files.reduce((sum, f) => sum + f.file.size, 0)
                    )}
                  </Text>
                </HStack>
                <Table size="sm" variant="simple">
                  <Tbody>
                    {files.map((fileData, index) => (
                      <Tr key={index}>
                        <Td px={0} py={2}>
                          <VStack align="stretch" spacing={1}>
                            <HStack justify="space-between">
                              <HStack flex={1} minW={0}>
                                <Box>{getStatusIcon(fileData.status)}</Box>
                                <VStack align="start" spacing={0} flex={1} minW={0}>
                                  <Input
                                    value={fileData.title}
                                    onChange={(e) =>
                                      updateFileMetadata(
                                        index,
                                        'title',
                                        e.target.value
                                      )
                                    }
                                    size="sm"
                                    isDisabled={
                                      isUploading || fileData.status !== 'pending'
                                    }
                                    placeholder="Track title"
                                  />
                                  <Text fontSize="xs" color="gray.500">
                                    {formatFileSize(fileData.file.size)}
                                  </Text>
                                </VStack>
                              </HStack>
                              {!isUploading && fileData.status === 'pending' && (
                                <IconButton
                                  aria-label="Remove file"
                                  icon={<IoClose />}
                                  size="sm"
                                  variant="ghost"
                                  colorScheme="red"
                                  onClick={() => removeFile(index)}
                                />
                              )}
                            </HStack>
                            {fileData.status === 'uploading' && (
                              <Progress
                                value={fileData.progress}
                                size="xs"
                                colorScheme="brand"
                                hasStripe
                                isAnimated
                              />
                            )}
                            {fileData.status === 'error' && (
                              <Badge colorScheme="red" fontSize="xs">
                                {fileData.error}
                              </Badge>
                            )}
                            {fileData.status === 'success' && (
                              <Badge colorScheme="green" fontSize="xs">
                                Uploaded successfully
                              </Badge>
                            )}
                          </VStack>
                        </Td>
                      </Tr>
                    ))}
                  </Tbody>
                </Table>
              </Box>
            )}
          </VStack>
        </ModalBody>

        <ModalFooter>
          <Button
            variant="ghost"
            mr={3}
            onClick={handleClose}
            isDisabled={isUploading}
          >
            Cancel
          </Button>
          <Button
            colorScheme="brand"
            onClick={handleUploadAll}
            isLoading={isUploading}
            loadingText="Uploading..."
            isDisabled={files.length === 0}
            leftIcon={<IoCloudUpload />}
          >
            Upload All ({files.length})
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  )
}

export default BulkUploadModal

