import { useState, useRef, useEffect } from 'react'
import {
  Box,
  Input,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalBody,
  ModalHeader,
  InputGroup,
  InputLeftElement,
  Spinner,
  Text,
  VStack,
  HStack,
  Badge,
  Image,
  Button,
  ButtonGroup,
  Divider,
  Alert,
  AlertIcon,
  CloseButton,
  Icon,
  Flex,
} from '@chakra-ui/react'
import { useEventListener } from '@chakra-ui/hooks'
import { FaSearch } from 'react-icons/fa'
import { IoMusicalNotes } from 'react-icons/io5'
import { searchTracks, formatDuration } from '../utils/api'
import { SearchResult } from '../interfaces/tracks'

type SourceFilter = 'all' | 'local' | 'youtube' | 'soundcloud'

interface EnhancedSearchBarProps {
  onTrackSelect?: (track: SearchResult) => void
}

function EnhancedSearchBar({ onTrackSelect }: EnhancedSearchBarProps) {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<SearchResult[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [sourceFilter, setSourceFilter] = useState<SourceFilter>('all')
  const [hasSearched, setHasSearched] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const searchInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (isOpen && searchInputRef.current) {
      searchInputRef.current.focus()
    }
  }, [isOpen])

  const handleKeyPress = (event: KeyboardEvent) => {
    if (event.key === 'Escape') {
      onClose()
      setQuery('')
      setResults([])
      setHasSearched(false)
      setError(null)
    }
    if (event.key === 'Enter' && isOpen) {
      handleSearch()
    }
  }

  useEventListener('keydown', handleKeyPress)

  const handleSearch = async () => {
    if (!query.trim()) return

    setLoading(true)
    setError(null)
    setHasSearched(true)

    try {
      const source = sourceFilter === 'all' ? undefined : sourceFilter
      const response = await searchTracks(query, source, 30)
      setResults(response.results)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Search failed')
      setResults([])
    } finally {
      setLoading(false)
    }
  }

  const handleSourceChange = (source: SourceFilter) => {
    setSourceFilter(source)
    if (hasSearched) {
      // Re-search with new filter
      setTimeout(() => handleSearch(), 100)
    }
  }

  const handleTrackClick = (track: SearchResult) => {
    if (onTrackSelect) {
      onTrackSelect(track)
    }
    onClose()
    setQuery('')
    setResults([])
    setHasSearched(false)
  }

  const getSourceColor = (source: string) => {
    switch (source) {
      case 'local':
        return 'green'
      case 'youtube':
        return 'red'
      case 'soundcloud':
        return 'orange'
      default:
        return 'gray'
    }
  }

  // Get background colors based on source
  const getSourceBg = (source: string) => {
    switch (source) {
      case 'local':
        return { bg: 'green.50', borderColor: 'green.200', iconColor: 'green.500' }
      case 'youtube':
        return { bg: 'red.50', borderColor: 'red.200', iconColor: 'red.500' }
      case 'soundcloud':
        return { bg: 'orange.50', borderColor: 'orange.200', iconColor: 'orange.500' }
      default:
        return { bg: 'blue.50', borderColor: 'blue.200', iconColor: 'blue.500' }
    }
  }

  // Thumbnail component with proper fallback
  const TrackThumbnail = ({ track }: { track: SearchResult }) => {
    const [imageError, setImageError] = useState(false)
    const hasThumbnail = track.thumbnail && track.thumbnail !== '' && !imageError
    const colors = getSourceBg(track.source)

    if (hasThumbnail) {
      return (
        <Image
          src={track.thumbnail}
          alt={track.title}
          boxSize="50px"
          minW="50px"
          borderRadius="md"
          objectFit="cover"
          onError={() => setImageError(true)}
          bg="gray.100"
        />
      )
    }

    // Fallback with icon - color coded by source
    return (
      <Flex
        boxSize="50px"
        minW="50px"
        borderRadius="md"
        bg={colors.bg}
        align="center"
        justify="center"
        border="1px solid"
        borderColor={colors.borderColor}
      >
        <Icon as={IoMusicalNotes} boxSize="24px" color={colors.iconColor} />
      </Flex>
    )
  }

  return (
    <Box textAlign="center" w="100%">
      <InputGroup>
        <InputLeftElement pointerEvents="none">
          <FaSearch color="gray.500" />
        </InputLeftElement>
        <Input
          ref={inputRef}
          placeholder="Search tracks (YouTube, SoundCloud, Local)"
          onClick={() => {
            onOpen()
            inputRef.current?.blur()
          }}
          onChange={() => ({})}
          value={''}
        />
      </InputGroup>

      <Modal isOpen={isOpen} onClose={onClose} size="xl">
        <ModalOverlay />
        <ModalContent maxH="80vh">
          <ModalHeader pb={2}>
            <VStack align="stretch" spacing={3}>
              <HStack>
                <Text flex={1}>Search Music</Text>
                <CloseButton onClick={onClose} />
              </HStack>

              <InputGroup size="md">
                <InputLeftElement pointerEvents="none">
                  <FaSearch color="gray.500" />
                </InputLeftElement>
                <Input
                  ref={searchInputRef}
                  placeholder="Type to search..."
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                  focusBorderColor="blue.500"
                />
              </InputGroup>

              <ButtonGroup size="sm" variant="outline" spacing={2}>
                <Button
                  colorScheme={sourceFilter === 'all' ? 'blue' : 'gray'}
                  onClick={() => handleSourceChange('all')}
                >
                  All Sources
                </Button>
                <Button
                  colorScheme={sourceFilter === 'local' ? 'green' : 'gray'}
                  onClick={() => handleSourceChange('local')}
                >
                  Local
                </Button>
                <Button
                  colorScheme={sourceFilter === 'youtube' ? 'red' : 'gray'}
                  onClick={() => handleSourceChange('youtube')}
                >
                  YouTube
                </Button>
                <Button
                  colorScheme={sourceFilter === 'soundcloud' ? 'orange' : 'gray'}
                  onClick={() => handleSourceChange('soundcloud')}
                >
                  SoundCloud
                </Button>
              </ButtonGroup>

              <Button
                colorScheme="blue"
                size="sm"
                onClick={handleSearch}
                isLoading={loading}
                isDisabled={!query.trim()}
              >
                Search
              </Button>
            </VStack>
          </ModalHeader>

          <Divider />

          <ModalBody overflowY="auto" maxH="50vh">
            {loading && (
              <VStack py={8} spacing={3}>
                <Spinner size="xl" color="blue.500" />
                <Text color="gray.600">Searching...</Text>
              </VStack>
            )}

            {error && (
              <Alert status="error" borderRadius="md">
                <AlertIcon />
                {error}
              </Alert>
            )}

            {!loading && hasSearched && results.length === 0 && !error && (
              <VStack py={8} spacing={2}>
                <Text fontSize="lg" color="gray.500">
                  No results found
                </Text>
                <Text fontSize="sm" color="gray.400">
                  Try a different search term or source
                </Text>
              </VStack>
            )}

            {!loading && results.length > 0 && (
              <VStack spacing={2} align="stretch" py={2}>
                {results.map((track) => (
                  <Box
                    key={`${track.source}-${track.id}`}
                    p={3}
                    borderWidth="1px"
                    borderRadius="md"
                    cursor="pointer"
                    transition="all 0.2s"
                    _hover={{ bg: 'gray.50', shadow: 'md', transform: 'translateY(-1px)' }}
                    onClick={() => handleTrackClick(track)}
                    bg="white"
                  >
                    <HStack spacing={3} align="center">
                      <TrackThumbnail track={track} />

                      <VStack align="start" flex={1} spacing={1} minW={0}>
                        <Text fontWeight="bold" fontSize="sm" noOfLines={1} w="100%">
                          {track.title}
                        </Text>
                        <Text fontSize="xs" color="gray.600" noOfLines={1} w="100%">
                          {track.artist}
                        </Text>
                        <HStack spacing={2} flexWrap="wrap">
                          <Badge 
                            colorScheme={getSourceColor(track.source)}
                            textTransform="uppercase"
                            fontSize="xx-small"
                            px={2}
                            py={0.5}
                          >
                            {track.source}
                          </Badge>
                          {track.duration > 0 && (
                            <Text fontSize="xs" color="gray.500" fontWeight="medium">
                              {formatDuration(track.duration)}
                            </Text>
                          )}
                        </HStack>
                      </VStack>
                    </HStack>
                  </Box>
                ))}

                <Text fontSize="xs" color="gray.500" textAlign="center" pt={2}>
                  {results.length} results found
                </Text>
              </VStack>
            )}

            {!loading && !hasSearched && (
              <VStack py={8} spacing={2}>
                <FaSearch size={48} color="gray.300" />
                <Text fontSize="md" color="gray.500">
                  Start typing to search
                </Text>
                <Text fontSize="sm" color="gray.400">
                  Search across local, YouTube, and SoundCloud
                </Text>
              </VStack>
            )}
          </ModalBody>
        </ModalContent>
      </Modal>
    </Box>
  )
}

export default EnhancedSearchBar

