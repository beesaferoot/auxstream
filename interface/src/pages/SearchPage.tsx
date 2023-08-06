import SearchBar from '../components/SearchBar.tsx'
import { AbsoluteCenter, Box, Center, Divider, Spinner } from '@chakra-ui/react'
import TrendingTracks from '../components/TrendingTracks.tsx'
import { getTrendingTracks } from '../utils/api.ts'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'

const SearchPage = () => {
  const tenMinutes = 600000
  const pageSize = 20
  const [page, setPage] = useState(1)
  const { isLoading, data } = useQuery({
    queryKey: ['getTrendingTracks'],
    queryFn: async () => getTrendingTracks(pageSize, page),
    staleTime: tenMinutes,
  })

  return (
    <Box className={'m-10'} position={'relative'}>
      <SearchBar />
      <Box position="relative" padding="10">
        <Divider />
        <AbsoluteCenter bg="white" px="4" fontWeight={'light'}>
          Trending
        </AbsoluteCenter>
      </Box>
      {data?.data && <TrendingTracks tracks={data.data} />}
      <Center>{isLoading && <Spinner />}</Center>
    </Box>
  )
}

export default SearchPage
