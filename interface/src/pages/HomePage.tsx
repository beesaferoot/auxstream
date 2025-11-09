import {
  Box,
  Container,
  Heading,
  Text,
  VStack,
  HStack,
  Button,
  SimpleGrid,
  Icon,
  useColorModeValue,
} from '@chakra-ui/react'
import { IoMusicalNotes, IoSearch, IoCloudDownload, IoPlay } from 'react-icons/io5'
import { useNavigate } from 'react-router-dom'

const HomePage = () => {
  const navigate = useNavigate()
  const cardBg = useColorModeValue('white', 'gray.800')

  const features = [
    {
      icon: IoSearch,
      title: 'Multi-Source Search',
      description: 'Search across YouTube, SoundCloud, and your local library simultaneously.',
    },
    {
      icon: IoPlay,
      title: 'Seamless Playback',
      description: 'High-quality audio streaming with an intuitive and beautiful player.',
    },
    {
      icon: IoCloudDownload,
      title: 'Multiple Sources',
      description: 'Access music from various platforms all in one convenient location.',
    },
    {
      icon: IoMusicalNotes,
      title: 'Trending Tracks',
      description: 'Discover what\'s popular and explore new music every day.',
    },
  ]

  return (
    <Box>
      {/* Hero Section */}
      <Box
        bgGradient="linear(to-br, blue.500, purple.600)"
        color="white"
        py={20}
        position="relative"
        overflow="hidden"
      >
        <Container maxW="container.xl">
          <VStack spacing={6} textAlign="center" maxW="3xl" mx="auto">
            <HStack spacing={3}>
              <Icon as={IoMusicalNotes} boxSize={16} />
            </HStack>
            <Heading
              fontSize={{ base: '4xl', md: '5xl', lg: '6xl' }}
              fontWeight="extrabold"
              lineHeight="shorter"
            >
              Your Music, All in One Place
            </Heading>
            <Text fontSize={{ base: 'lg', md: 'xl' }} opacity={0.9} maxW="2xl">
              Stream from YouTube, SoundCloud, and your local library with AuxStream.
              The ultimate music streaming hub for all your favorite tracks.
            </Text>
            <HStack spacing={4} pt={4}>
              <Button
                size="lg"
                colorScheme="whiteAlpha"
                bg="white"
                color="blue.600"
                leftIcon={<IoSearch />}
                onClick={() => navigate('/player')}
                _hover={{
                  transform: 'translateY(-2px)',
                  shadow: 'xl',
                }}
                transition="all 0.2s"
              >
                Start Searching
              </Button>
              <Button
                size="lg"
                variant="outline"
                color="white"
                borderColor="white"
                _hover={{
                  bg: 'whiteAlpha.200',
                }}
                onClick={() => navigate('/trending')}
              >
                View Trending
              </Button>
            </HStack>
          </VStack>
        </Container>
      </Box>

      {/* Features Section */}
      <Container maxW="container.xl" py={20}>
        <VStack spacing={12}>
          <VStack spacing={4} textAlign="center">
            <Heading size="xl">Why Choose AuxStream?</Heading>
            <Text fontSize="lg" color="gray.600" maxW="2xl">
              Experience the best of music streaming with our powerful features and
              beautiful interface.
            </Text>
          </VStack>

          <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={8} w="100%">
            {features.map((feature, index) => (
              <Box
                key={index}
                bg={cardBg}
                p={8}
                borderRadius="xl"
                shadow="md"
                transition="all 0.3s"
                _hover={{
                  transform: 'translateY(-8px)',
                  shadow: 'xl',
                }}
              >
                <VStack spacing={4} align="start">
                  <Icon
                    as={feature.icon}
                    boxSize={12}
                    color="blue.500"
                    p={2}
                    bg="blue.50"
                    borderRadius="lg"
                  />
                  <Heading size="md">{feature.title}</Heading>
                  <Text color="gray.600">{feature.description}</Text>
                </VStack>
              </Box>
            ))}
          </SimpleGrid>
        </VStack>
      </Container>

      {/* CTA Section */}
      <Box bg="gray.50" py={20}>
        <Container maxW="container.md">
          <VStack spacing={6} textAlign="center">
            <Heading size="xl">Ready to Start Streaming?</Heading>
            <Text fontSize="lg" color="gray.600">
              Jump into the world of unlimited music streaming right now.
            </Text>
            <Button
              size="lg"
              colorScheme="brand"
              leftIcon={<IoPlay />}
              onClick={() => navigate('/player')}
              _hover={{
                transform: 'translateY(-2px)',
                shadow: 'lg',
              }}
              transition="all 0.2s"
            >
              Launch Player
            </Button>
          </VStack>
        </Container>
      </Box>
    </Box>
  )
}

export default HomePage

