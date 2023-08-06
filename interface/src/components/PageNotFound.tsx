import { Text, Box, AbsoluteCenter } from "@chakra-ui/react"

const PageNotFound = () => {
  return (
    <Box position="relative" h="100px">
      <AbsoluteCenter>
        <Text fontSize="6sm">404 Page not Found</Text>
      </AbsoluteCenter>
    </Box>
  )
}

export default PageNotFound
