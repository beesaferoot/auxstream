import { useState } from 'react'
import {
  Box,
  Container,
  VStack,
  Heading,
  Text,
  Divider,
  HStack,
  Badge,
  Button,
  useDisclosure,
  Tooltip,
} from '@chakra-ui/react'
import { IoCloudUpload, IoLogIn } from 'react-icons/io5'
import EnhancedSearchBar from '../components/EnhancedSearchBar'
import { SearchResult } from '../interfaces/tracks'
import EnhancedAudioPlayer from '../components/EnhancedAudioPlayer'
import { getAudioUrl, trackPlay } from '../utils/api'
import UploadTrackModal from '../components/UploadTrackModal'
import BulkUploadModal from '../components/BulkUploadModal'
import AuthModal from '../components/AuthModal'
import { useAuth } from '../context/AuthContext'

function MusicPlayerPage() {
  const [selectedTrack, setSelectedTrack] = useState<SearchResult | null>(null)
  const [recentSearches, setRecentSearches] = useState<SearchResult[]>([])
  const { isOpen: isUploadOpen, onOpen: onUploadOpen, onClose: onUploadClose } = useDisclosure()
  const { isOpen: isBulkUploadOpen, onOpen: onBulkUploadOpen, onClose: onBulkUploadClose } = useDisclosure()
  const { isOpen: isAuthOpen, onOpen: onAuthOpen, onClose: onAuthClose } = useDisclosure()
  const { isAuthenticated } = useAuth()

  const handleTrackSelect = (track: SearchResult) => {
    setSelectedTrack(track)

    // Track the play event (only for local tracks with valid UUIDs)
    if (track.source === 'local') {
      trackPlay(track.id)
    }

    // Add to recent searches (keep last 5)
    setRecentSearches((prev) => {
      const filtered = prev.filter(
        (t) => t.id !== track.id || t.source !== track.source
      )
      return [track, ...filtered].slice(0, 5)
    })
  }

  return (
    <Box minH="calc(100vh - 64px)" bg="gray.50" py={8}>
      <Container maxW="container.xl">
        <VStack spacing={8} align="stretch">
          {/* Header */}
          <VStack spacing={3} align="center" py={6}>
            <Heading
              size="2xl"
              bgGradient="linear(to-r, blue.500, purple.600)"
              bgClip="text"
            >
              Music Player
            </Heading>
            <Text color="gray.600" fontSize="lg" textAlign="center">
              Search and stream from YouTube, SoundCloud, and Local library
            </Text>
            <HStack spacing={3}>
              <Badge colorScheme="green" fontSize="sm" px={3} py={1} borderRadius="full">
                Local
              </Badge>
              <Badge colorScheme="red" fontSize="sm" px={3} py={1} borderRadius="full">
                YouTube
              </Badge>
              <Badge colorScheme="orange" fontSize="sm" px={3} py={1} borderRadius="full">
                SoundCloud
              </Badge>
            </HStack>
          </VStack>

          {/* Audio Player */}
          <Box>
            <EnhancedAudioPlayer
              src={selectedTrack ? getAudioUrl(selectedTrack.stream_url) : undefined}
              title={selectedTrack?.title}
              artist={selectedTrack?.artist}
              thumbnail={selectedTrack?.thumbnail}
            />
          </Box>

          <Divider />

          {/* Search Bar */}
          <Box>
            <HStack justify="space-between" align="center" mb={4}>
              <Heading size="md">Search Music</Heading>
              {isAuthenticated ? (
                <HStack>
                  <Button
                    leftIcon={<IoCloudUpload />}
                    colorScheme="brand"
                    size="sm"
                    onClick={onUploadOpen}
                  >
                    Upload Track
                  </Button>
                  <Button
                    leftIcon={<IoCloudUpload />}
                    colorScheme="brand"
                    variant="outline"
                    size="sm"
                    onClick={onBulkUploadOpen}
                  >
                    Bulk Upload
                  </Button>
                </HStack>
              ) : (
                <Tooltip
                  label="Please login to upload tracks"
                  placement="left"
                  hasArrow
                >
                  <Button
                    leftIcon={<IoLogIn />}
                    colorScheme="brand"
                    variant="outline"
                    size="sm"
                    onClick={onAuthOpen}
                  >
                    Login to Upload
                  </Button>
                </Tooltip>
              )}
            </HStack>
            <EnhancedSearchBar onTrackSelect={handleTrackSelect} />
          </Box>

        {/* Recent Searches */}
        {recentSearches.length > 0 && (
          <>
            <Divider />
            <Box>
              <Heading size="md" mb={3}>
                Recent Searches
              </Heading>
              <VStack spacing={2} align="stretch">
                {recentSearches.map((track, index) => (
                  <Box
                    key={`${track.source}-${track.id}-${index}`}
                    p={3}
                    borderWidth="1px"
                    borderRadius="md"
                    cursor="pointer"
                    bg={selectedTrack?.id === track.id ? 'blue.50' : 'white'}
                    _hover={{ bg: 'gray.50', shadow: 'sm' }}
                    onClick={() => handleTrackSelect(track)}
                  >
                    <HStack justify="space-between">
                      <VStack align="start" spacing={0}>
                        <Text fontWeight="bold" fontSize="sm">
                          {track.title}
                        </Text>
                        <Text fontSize="xs" color="gray.600">
                          {track.artist}
                        </Text>
                      </VStack>
                      <Badge
                        colorScheme={
                          track.source === 'local'
                            ? 'green'
                            : track.source === 'youtube'
                              ? 'red'
                              : 'orange'
                        }
                      >
                        {track.source}
                      </Badge>
                    </HStack>
                  </Box>
                ))}
              </VStack>
            </Box>
          </>
        )}

        {/* Footer Info */}
        <Divider />
        <VStack spacing={2}>
          <Text fontSize="sm" color="gray.600" textAlign="center">
            Music Player Page
          </Text>
        </VStack>
      </VStack>
    </Container>

    {/* Upload Modals */}
    <UploadTrackModal isOpen={isUploadOpen} onClose={onUploadClose} />
    <BulkUploadModal isOpen={isBulkUploadOpen} onClose={onBulkUploadClose} />

    {/* Auth Modal */}
    <AuthModal isOpen={isAuthOpen} onClose={onAuthClose} />
    </Box>
  )
}

export default MusicPlayerPage
