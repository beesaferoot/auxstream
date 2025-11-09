import { Text, Box, VStack, Button, Heading } from '@chakra-ui/react'
import { IoHome, IoSearchOutline } from 'react-icons/io5'
import { useNavigate } from 'react-router-dom'

const PageNotFound = () => {
  const navigate = useNavigate()

  return (
    <Box
      minH="calc(100vh - 64px)"
      display="flex"
      alignItems="center"
      justifyContent="center"
      bg="gray.50"
    >
      <VStack spacing={6} textAlign="center" p={8}>
        {/* Large 404 */}
        <Heading
          fontSize="9xl"
          fontWeight="extrabold"
          bgGradient="linear(to-r, blue.500, purple.600)"
          bgClip="text"
          lineHeight="1"
        >
          404
        </Heading>

        {/* Message */}
        <VStack spacing={2}>
          <Heading size="lg" color="gray.700">
            Page Not Found
          </Heading>
          <Text color="gray.500" fontSize="md" maxW="md">
            Oops! The page you're looking for doesn't exist. It might have been
            moved or deleted.
          </Text>
        </VStack>

        {/* Action Buttons */}
        <VStack spacing={3} pt={4}>
          <Button
            colorScheme="brand"
            size="lg"
            leftIcon={<IoHome />}
            onClick={() => navigate('/')}
            _hover={{ transform: 'translateY(-2px)', shadow: 'lg' }}
            transition="all 0.2s"
          >
            Go Home
          </Button>
          <Button
            variant="ghost"
            size="md"
            leftIcon={<IoSearchOutline />}
            onClick={() => navigate('/player')}
          >
            Search Music
          </Button>
        </VStack>
      </VStack>
    </Box>
  )
}

export default PageNotFound
