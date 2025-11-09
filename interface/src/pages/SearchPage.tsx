import {
  Box,
  Center,
  Spinner,
  Container,
  Heading,
  VStack,
  Text,
  HStack,
  Button,
  ButtonGroup,
} from '@chakra-ui/react'
import TrendingTracks from '../components/TrendingTracks.tsx'
import { getTrendingTracks } from '../utils/api.ts'
import { useQuery } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { IoTrendingUp, IoTime, IoMusicalNotes, IoChevronDown } from 'react-icons/io5'
import { Track } from '../interfaces/tracks'

type SortType = 'trending' | 'recent' | 'default'

const SearchPage = () => {
  const tenMinutes = 600000
  const pageSize = 20
  const [page, setPage] = useState(1)
  const [sortType, setSortType] = useState<SortType>('trending')
  const [allTracks, setAllTracks] = useState<Track[]>([])
  const [hasMore, setHasMore] = useState(true)
  
  const { isLoading, data, isFetching } = useQuery({
    queryKey: ['getTrendingTracks', page, sortType],
    queryFn: async () => getTrendingTracks(pageSize, page, sortType, 30),
    staleTime: tenMinutes,
  })

  // Update allTracks when new data arrives
  useEffect(() => {
    if (data?.data) {
      if (page === 1) {
        // First page - replace all tracks
        setAllTracks(data.data)
      } else {
        // Subsequent pages - append tracks
        setAllTracks(prev => [...prev, ...data.data])
      }
      
      // Check if we have more data
      setHasMore(data.data.length === pageSize)
    }
  }, [data, page, pageSize])

  const handleSortChange = (newSort: SortType) => {
    setSortType(newSort)
    setPage(1)
    setAllTracks([]) // Clear tracks when changing sort
    setHasMore(true)
  }

  const handleLoadMore = () => {
    if (!isFetching && hasMore) {
      setPage(prev => prev + 1)
    }
  }

  return (
    <Box minH="calc(100vh - 64px)" bg="gray.50" py={8}>
      <Container maxW="container.xl">
        <VStack spacing={8} align="stretch">
          {/* Header */}
          <VStack spacing={4} align="start">
            <HStack justify="space-between" w="100%">
              <VStack spacing={2} align="start">
                <Heading size="xl">
                  {sortType === 'trending' && 'Trending Tracks'}
                  {sortType === 'recent' && 'Recently Added'}
                  {sortType === 'default' && 'All Tracks'}
                </Heading>
                <Text color="gray.600" fontSize="lg">
                  {sortType === 'trending' && 'Discover the most popular music right now'}
                  {sortType === 'recent' && 'Check out the latest additions'}
                  {sortType === 'default' && 'Browse all available tracks'}
                </Text>
              </VStack>
            </HStack>

            {/* Sort Options */}
            <ButtonGroup size="sm" variant="outline" spacing={3}>
              <Button
                leftIcon={<IoTrendingUp />}
                colorScheme={sortType === 'trending' ? 'blue' : 'gray'}
                variant={sortType === 'trending' ? 'solid' : 'outline'}
                onClick={() => handleSortChange('trending')}
              >
                Trending
              </Button>
              <Button
                leftIcon={<IoTime />}
                colorScheme={sortType === 'recent' ? 'green' : 'gray'}
                variant={sortType === 'recent' ? 'solid' : 'outline'}
                onClick={() => handleSortChange('recent')}
              >
                Recent
              </Button>
              <Button
                leftIcon={<IoMusicalNotes />}
                colorScheme={sortType === 'default' ? 'purple' : 'gray'}
                variant={sortType === 'default' ? 'solid' : 'outline'}
                onClick={() => handleSortChange('default')}
              >
                All
              </Button>
            </ButtonGroup>
          </VStack>

          {/* Content */}
          {isLoading && page === 1 ? (
            <Center py={20}>
              <VStack spacing={4}>
                <Spinner size="xl" color="brand.500" thickness="4px" />
                <Text color="gray.600">Loading tracks...</Text>
              </VStack>
            </Center>
          ) : allTracks.length > 0 ? (
            <VStack spacing={6} align="stretch">
              <TrendingTracks tracks={allTracks} />
              
              {/* Load More Section */}
              {hasMore && (
                <Center py={8}>
                  <Button
                    size="lg"
                    colorScheme="brand"
                    variant="outline"
                    leftIcon={isFetching ? <Spinner size="sm" /> : <IoChevronDown />}
                    onClick={handleLoadMore}
                    isLoading={isFetching}
                    loadingText="Loading more..."
                    isDisabled={isFetching}
                  >
                    Load More Tracks
                  </Button>
                </Center>
              )}

              {/* End of results indicator */}
              {!hasMore && allTracks.length > 0 && (
                <Center py={8}>
                  <VStack spacing={2}>
                    <Text color="gray.500" fontSize="sm">
                      You've reached the end
                    </Text>
                    <Text color="gray.400" fontSize="xs">
                      Showing all {allTracks.length} tracks
                    </Text>
                  </VStack>
                </Center>
              )}
            </VStack>
          ) : (
            <Center py={20}>
              <VStack spacing={2}>
                <Text color="gray.500" fontSize="lg">
                  No tracks available at the moment
                </Text>
                <Text color="gray.400" fontSize="sm">
                  Try uploading some tracks to get started
                </Text>
              </VStack>
            </Center>
          )}
        </VStack>
      </Container>
    </Box>
  )
}

export default SearchPage
