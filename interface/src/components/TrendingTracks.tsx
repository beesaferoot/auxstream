import { useState } from 'react'
import {
  Box,
  Card,
  CardBody,
  CardFooter,
  Grid,
  Image,
  Text,
  VStack,
  HStack,
  IconButton,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalCloseButton,
  useToast,
  Flex,
  Icon,
} from '@chakra-ui/react'
import { Track } from '../interfaces/tracks'
import { IoPlay, IoHeart, IoHeartOutline, IoShareSocial, IoMusicalNotes } from 'react-icons/io5'
import { getAudioUrl, trackPlay } from '../utils/api'
import AudioPlayer from './AudioPlayer'

const TrendingTracks = ({ tracks }: { tracks: Track[] }) => {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const [selectedTrack, setSelectedTrack] = useState<Track | null>(null)
  const [likedTracks, setLikedTracks] = useState<Set<number>>(new Set())
  const [audioError, setAudioError] = useState(false)
  const [imageErrors, setImageErrors] = useState<Set<number>>(new Set())
  const toast = useToast()

  const handlePlayClick = (track: Track) => {
    setSelectedTrack(track)
    setAudioError(false)
    onOpen()
    
    // Track the play event
    trackPlay(track.id)
  }

  const handleAudioError = () => {
    setAudioError(true)
    toast({
      title: 'Audio Error',
      description: 'Failed to load audio file. Please try again.',
      status: 'error',
      duration: 3000,
      isClosable: true,
    })
  }

  const handleImageError = (index: number) => {
    setImageErrors(prev => new Set(prev).add(index))
  }

  const toggleLike = (trackIndex: number, e: React.MouseEvent) => {
    e.stopPropagation()
    setLikedTracks((prev) => {
      const newSet = new Set(prev)
      if (newSet.has(trackIndex)) {
        newSet.delete(trackIndex)
      } else {
        newSet.add(trackIndex)
        toast({
          title: 'Added to favorites',
          status: 'success',
          duration: 2000,
          isClosable: true,
        })
      }
      return newSet
    })
  }

  const handleShare = (track: Track, e: React.MouseEvent) => {
    e.stopPropagation()
    toast({
      title: 'Link copied to clipboard',
      description: `${track.title} - ${track.artist}`,
      status: 'info',
      duration: 2000,
      isClosable: true,
    })
  }

  return (
    <>
      <Grid
        templateColumns={{
          base: 'repeat(1, 1fr)',
          sm: 'repeat(2, 1fr)',
          md: 'repeat(3, 1fr)',
          lg: 'repeat(4, 1fr)',
          xl: 'repeat(5, 1fr)',
        }}
        gap={6}
      >
        {tracks.map((track, i) => {
          const isLiked = likedTracks.has(i)
          const hasImageError = imageErrors.has(i)
          const hasThumbnail = track.thumbnail && track.thumbnail !== '' && !hasImageError

          return (
            <Card
              key={i}
              bg="white"
              overflow="hidden"
              cursor="pointer"
              transition="all 0.3s"
              _hover={{
                transform: 'translateY(-8px)',
                shadow: 'xl',
              }}
              onClick={() => handlePlayClick(track)}
            >
              <CardBody p={0}>
                <Box position="relative" overflow="hidden" bg="gray.100">
                  {hasThumbnail ? (
                    <Image
                      src={track.thumbnail}
                      alt={track.title}
                      w="100%"
                      h="200px"
                      objectFit="cover"
                      onError={() => handleImageError(i)}
                    />
                  ) : (
                    <Flex
                      w="100%"
                      h="200px"
                      bg="blue.50"
                      align="center"
                      justify="center"
                    >
                      <Icon as={IoMusicalNotes} boxSize="80px" color="blue.300" />
                    </Flex>
                  )}
                  <Box
                    position="absolute"
                    top={0}
                    left={0}
                    right={0}
                    bottom={0}
                    bg="blackAlpha.600"
                    opacity={0}
                    transition="opacity 0.3s"
                    _groupHover={{ opacity: 1 }}
                    display="flex"
                    alignItems="center"
                    justifyContent="center"
                  >
                    <IconButton
                      aria-label="Play track"
                      icon={<IoPlay size={32} />}
                      colorScheme="brand"
                      size="lg"
                      borderRadius="full"
                      _hover={{ transform: 'scale(1.1)' }}
                    />
                  </Box>
                </Box>

                <VStack align="stretch" p={4} spacing={2}>
                  <Text
                    fontWeight="bold"
                    fontSize="md"
                    noOfLines={1}
                    title={track.title}
                  >
                    {track.title}
                  </Text>
                  <Text fontSize="sm" color="gray.600" noOfLines={1}>
                    {track.artist?.name || 'Unknown Artist'}
                  </Text>
                </VStack>
              </CardBody>

              <CardFooter pt={0} px={4} pb={4}>
                <HStack w="100%" justify="space-between">
                  <IconButton
                    aria-label="Like"
                    icon={isLiked ? <IoHeart /> : <IoHeartOutline />}
                    variant="ghost"
                    colorScheme={isLiked ? 'red' : 'gray'}
                    size="sm"
                    onClick={(e) => toggleLike(i, e)}
                  />
                  <IconButton
                    aria-label="Share"
                    icon={<IoShareSocial />}
                    variant="ghost"
                    colorScheme="gray"
                    size="sm"
                    onClick={(e) => handleShare(track, e)}
                  />
                </HStack>
              </CardFooter>
            </Card>
          )
        })}
      </Grid>

      {/* Playback Modal */}
      <Modal isOpen={isOpen} onClose={onClose} size="xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Now Playing</ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            {selectedTrack && (
              <VStack spacing={4}>
                {selectedTrack.thumbnail ? (
                  <Image
                    src={selectedTrack.thumbnail}
                    alt={selectedTrack.title}
                    borderRadius="lg"
                    maxH="300px"
                    objectFit="cover"
                  />
                ) : (
                  <Flex
                    w="100%"
                    h="300px"
                    borderRadius="lg"
                    bg="blue.50"
                    align="center"
                    justify="center"
                  >
                    <Icon as={IoMusicalNotes} boxSize="120px" color="blue.300" />
                  </Flex>
                )}
                <VStack spacing={1} w="100%">
                  <Text fontWeight="bold" fontSize="xl">
                    {selectedTrack.title}
                  </Text>
                  <Text color="gray.600">{selectedTrack.artist?.name || 'Unknown Artist'}</Text>
                </VStack>
                {audioError && (
                  <Text color="red.500" fontSize="sm" textAlign="center">
                    Unable to load audio. The file may not exist or there may be a connection issue.
                  </Text>
                )}
                <Box w="100%">
                  <AudioPlayer 
                    src={getAudioUrl(selectedTrack.file)}
                    onError={handleAudioError}
                  />
                </Box>
              </VStack>
            )}
          </ModalBody>
        </ModalContent>
      </Modal>
    </>
  )
}

export default TrendingTracks
