import {
  Box,
  Container,
  VStack,
  Heading,
  Text,
  Card,
  CardBody,
  HStack,
  Avatar,
  Badge,
} from '@chakra-ui/react'
import { useAuth } from '../context/AuthContext'

const ProfilePage = () => {
  const { userEmail, isAuthenticated } = useAuth()

  if (!isAuthenticated) {
    return (
      <Box minH="calc(100vh - 64px)" bg="gray.50" py={8}>
        <Container maxW="container.md">
          <VStack spacing={6} align="center" py={20}>
            <Heading size="lg">Please log in to view your profile</Heading>
            <Text color="gray.600">
              You need to be authenticated to access this page.
            </Text>
          </VStack>
        </Container>
      </Box>
    )
  }

  return (
    <Box minH="calc(100vh - 64px)" bg="gray.50" py={8}>
      <Container maxW="container.md">
        <VStack spacing={8} align="stretch">
          <Heading size="xl">Profile</Heading>

          <Card>
            <CardBody>
              <VStack spacing={6} align="stretch">
                <HStack spacing={4}>
                  <Avatar
                    size="xl"
                    name={userEmail || 'User'}
                    bg="brand.500"
                  />
                  <VStack align="start" spacing={1}>
                    <Heading size="md">{userEmail}</Heading>
                    <Badge colorScheme="green">Active</Badge>
                  </VStack>
                </HStack>

                <Box>
                  <Text fontWeight="bold" mb={2}>
                    Account Information
                  </Text>
                  <VStack align="stretch" spacing={2}>
                    <HStack justify="space-between">
                      <Text color="gray.600">Email:</Text>
                      <Text fontWeight="medium">{userEmail}</Text>
                    </HStack>
                    <HStack justify="space-between">
                      <Text color="gray.600">Account Type:</Text>
                      <Text fontWeight="medium">Free</Text>
                    </HStack>
                  </VStack>
                </Box>

                <Box>
                  <Text fontWeight="bold" mb={2}>
                    Statistics
                  </Text>
                  <VStack align="stretch" spacing={2}>
                    <HStack justify="space-between">
                      <Text color="gray.600">Tracks Uploaded:</Text>
                      <Text fontWeight="medium">0</Text>
                    </HStack>
                    <HStack justify="space-between">
                      <Text color="gray.600">Favorites:</Text>
                      <Text fontWeight="medium">0</Text>
                    </HStack>
                  </VStack>
                </Box>
              </VStack>
            </CardBody>
          </Card>

          <Text fontSize="sm" color="gray.500" textAlign="center">
            Profile features are coming soon!
          </Text>
        </VStack>
      </Container>
    </Box>
  )
}

export default ProfilePage

