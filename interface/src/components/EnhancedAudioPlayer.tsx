import { useState, useRef, useEffect } from 'react'
import {
  Box,
  Flex,
  IconButton,
  Slider,
  SliderTrack,
  SliderFilledTrack,
  SliderThumb,
  Text,
  HStack,
  VStack,
  Image,
  useColorModeValue,
} from '@chakra-ui/react'
import {
  IoPlay,
  IoPause,
  IoPlaySkipForward,
  IoPlaySkipBack,
  IoVolumeHigh,
  IoVolumeMute,
  IoRepeat,
  IoShuffle,
  IoMusicalNotes,
} from 'react-icons/io5'

interface EnhancedAudioPlayerProps {
  src?: string
  title?: string
  artist?: string
  thumbnail?: string
  onNext?: () => void
  onPrevious?: () => void
}

const EnhancedAudioPlayer = ({
  src,
  title = 'No track selected',
  artist = 'Unknown artist',
  thumbnail,
  onNext,
  onPrevious,
}: EnhancedAudioPlayerProps) => {
  const audioRef = useRef<HTMLAudioElement>(null)
  const [isPlaying, setIsPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const [volume, setVolume] = useState(1)
  const [isMuted, setIsMuted] = useState(false)
  const [isRepeat, setIsRepeat] = useState(false)
  const [isShuffle, setIsShuffle] = useState(false)

  const bg = useColorModeValue('white', 'gray.800')
  const borderColor = useColorModeValue('gray.200', 'gray.700')
  const iconBg = useColorModeValue('gray.100', 'gray.700')
  const iconColor = useColorModeValue('gray.400', 'gray.500')

  useEffect(() => {
    const audio = audioRef.current
    if (!audio) return

    const handleTimeUpdate = () => setCurrentTime(audio.currentTime)
    const handleDurationChange = () => setDuration(audio.duration)
    const handleEnded = () => {
      if (isRepeat) {
        audio.currentTime = 0
        audio.play()
      } else {
        setIsPlaying(false)
        if (onNext) onNext()
      }
    }

    audio.addEventListener('timeupdate', handleTimeUpdate)
    audio.addEventListener('durationchange', handleDurationChange)
    audio.addEventListener('ended', handleEnded)

    return () => {
      audio.removeEventListener('timeupdate', handleTimeUpdate)
      audio.removeEventListener('durationchange', handleDurationChange)
      audio.removeEventListener('ended', handleEnded)
    }
  }, [isRepeat, onNext])

  useEffect(() => {
    if (src && audioRef.current) {
      audioRef.current.load()
      if (isPlaying) {
        audioRef.current.play()
      }
    }
  }, [src])

  const togglePlay = () => {
    if (!audioRef.current || !src) return

    if (isPlaying) {
      audioRef.current.pause()
    } else {
      audioRef.current.play()
    }
    setIsPlaying(!isPlaying)
  }

  const handleSeek = (value: number) => {
    if (!audioRef.current) return
    audioRef.current.currentTime = value
    setCurrentTime(value)
  }

  const handleVolumeChange = (value: number) => {
    if (!audioRef.current) return
    audioRef.current.volume = value
    setVolume(value)
    setIsMuted(value === 0)
  }

  const toggleMute = () => {
    if (!audioRef.current) return
    if (isMuted) {
      audioRef.current.volume = volume || 0.5
      setIsMuted(false)
    } else {
      audioRef.current.volume = 0
      setIsMuted(true)
    }
  }

  const formatTime = (time: number) => {
    if (isNaN(time)) return '0:00'
    const minutes = Math.floor(time / 60)
    const seconds = Math.floor(time % 60)
    return `${minutes}:${seconds.toString().padStart(2, '0')}`
  }

  return (
    <Box
      bg={bg}
      borderWidth="1px"
      borderColor={borderColor}
      borderRadius="xl"
      p={6}
      shadow="lg"
      w="100%"
    >
      <audio ref={audioRef} src={src} />

      <Flex gap={6} align="center">
        {/* Track Info with Thumbnail */}
        <HStack spacing={4} flex={1} minW={0}>
          {thumbnail ? (
            <Image
              src={thumbnail}
              alt={title}
              boxSize="60px"
              borderRadius="md"
              objectFit="cover"
              fallback={
                <Box
                  boxSize="60px"
                  borderRadius="md"
                  bg={iconBg}
                  display="flex"
                  alignItems="center"
                  justifyContent="center"
                >
                  <Box as={IoMusicalNotes} boxSize="32px" color={iconColor} />
                </Box>
              }
            />
          ) : (
            <Box
              boxSize="60px"
              borderRadius="md"
              bg={iconBg}
              display="flex"
              alignItems="center"
              justifyContent="center"
            >
              <Box as={IoMusicalNotes} boxSize="32px" color={iconColor} />
            </Box>
          )}
          <VStack align="start" spacing={0} flex={1} minW={0}>
            <Text fontWeight="bold" fontSize="md" noOfLines={1}>
              {title}
            </Text>
            <Text fontSize="sm" color="gray.500" noOfLines={1}>
              {artist}
            </Text>
          </VStack>
        </HStack>

        {/* Controls */}
        <VStack spacing={3} flex={2}>
          {/* Playback Controls */}
          <HStack spacing={2}>
            <IconButton
              aria-label="Shuffle"
              icon={<IoShuffle />}
              variant="ghost"
              size="sm"
              colorScheme={isShuffle ? 'brand' : 'gray'}
              onClick={() => setIsShuffle(!isShuffle)}
              isDisabled={!src}
            />
            <IconButton
              aria-label="Previous"
              icon={<IoPlaySkipBack />}
              variant="ghost"
              onClick={onPrevious}
              isDisabled={!onPrevious || !src}
            />
            <IconButton
              aria-label={isPlaying ? 'Pause' : 'Play'}
              icon={isPlaying ? <IoPause size={24} /> : <IoPlay size={24} />}
              colorScheme="brand"
              size="lg"
              borderRadius="full"
              onClick={togglePlay}
              isDisabled={!src}
            />
            <IconButton
              aria-label="Next"
              icon={<IoPlaySkipForward />}
              variant="ghost"
              onClick={onNext}
              isDisabled={!onNext || !src}
            />
            <IconButton
              aria-label="Repeat"
              icon={<IoRepeat />}
              variant="ghost"
              size="sm"
              colorScheme={isRepeat ? 'brand' : 'gray'}
              onClick={() => setIsRepeat(!isRepeat)}
              isDisabled={!src}
            />
          </HStack>

          {/* Progress Bar */}
          <HStack w="100%" spacing={3}>
            <Text fontSize="xs" color="gray.500" minW="40px">
              {formatTime(currentTime)}
            </Text>
            <Slider
              aria-label="seek"
              value={currentTime}
              min={0}
              max={duration || 0}
              onChange={handleSeek}
              isDisabled={!src}
              colorScheme="brand"
            >
              <SliderTrack>
                <SliderFilledTrack />
              </SliderTrack>
              <SliderThumb />
            </Slider>
            <Text fontSize="xs" color="gray.500" minW="40px">
              {formatTime(duration)}
            </Text>
          </HStack>
        </VStack>

        {/* Volume Control */}
        <HStack spacing={2} w="150px">
          <IconButton
            aria-label="Mute"
            icon={isMuted ? <IoVolumeMute /> : <IoVolumeHigh />}
            variant="ghost"
            size="sm"
            onClick={toggleMute}
            isDisabled={!src}
          />
          <Slider
            aria-label="volume"
            value={isMuted ? 0 : volume}
            min={0}
            max={1}
            step={0.01}
            onChange={handleVolumeChange}
            isDisabled={!src}
            colorScheme="brand"
          >
            <SliderTrack>
              <SliderFilledTrack />
            </SliderTrack>
            <SliderThumb />
          </Slider>
        </HStack>
      </Flex>
    </Box>
  )
}

export default EnhancedAudioPlayer

