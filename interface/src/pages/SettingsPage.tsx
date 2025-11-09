import {
  Box,
  Container,
  VStack,
  Heading,
  Text,
  Card,
  CardBody,
  FormControl,
  FormLabel,
  Switch,
  Divider,
} from '@chakra-ui/react'
import { useAuth } from '../context/AuthContext'

const SettingsPage = () => {
  const { isAuthenticated } = useAuth()

  if (!isAuthenticated) {
    return (
      <Box minH="calc(100vh - 64px)" bg="gray.50" py={8}>
        <Container maxW="container.md">
          <VStack spacing={6} align="center" py={20}>
            <Heading size="lg">Please log in to access settings</Heading>
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
          <Heading size="xl">Settings</Heading>

          <Card>
            <CardBody>
              <VStack spacing={6} align="stretch">
                <Box>
                  <Text fontWeight="bold" fontSize="lg" mb={4}>
                    Playback Settings
                  </Text>
                  <VStack spacing={4} align="stretch">
                    <FormControl display="flex" alignItems="center">
                      <FormLabel htmlFor="autoplay" mb="0" flex={1}>
                        Autoplay next track
                      </FormLabel>
                      <Switch id="autoplay" colorScheme="brand" />
                    </FormControl>
                    <FormControl display="flex" alignItems="center">
                      <FormLabel htmlFor="quality" mb="0" flex={1}>
                        High quality audio
                      </FormLabel>
                      <Switch id="quality" colorScheme="brand" defaultChecked />
                    </FormControl>
                  </VStack>
                </Box>

                <Divider />

                <Box>
                  <Text fontWeight="bold" fontSize="lg" mb={4}>
                    Notifications
                  </Text>
                  <VStack spacing={4} align="stretch">
                    <FormControl display="flex" alignItems="center">
                      <FormLabel htmlFor="email-notif" mb="0" flex={1}>
                        Email notifications
                      </FormLabel>
                      <Switch id="email-notif" colorScheme="brand" />
                    </FormControl>
                    <FormControl display="flex" alignItems="center">
                      <FormLabel htmlFor="browser-notif" mb="0" flex={1}>
                        Browser notifications
                      </FormLabel>
                      <Switch id="browser-notif" colorScheme="brand" />
                    </FormControl>
                  </VStack>
                </Box>

                <Divider />

                <Box>
                  <Text fontWeight="bold" fontSize="lg" mb={4}>
                    Privacy
                  </Text>
                  <VStack spacing={4} align="stretch">
                    <FormControl display="flex" alignItems="center">
                      <FormLabel htmlFor="public-profile" mb="0" flex={1}>
                        Public profile
                      </FormLabel>
                      <Switch id="public-profile" colorScheme="brand" />
                    </FormControl>
                    <FormControl display="flex" alignItems="center">
                      <FormLabel htmlFor="show-activity" mb="0" flex={1}>
                        Show listening activity
                      </FormLabel>
                      <Switch
                        id="show-activity"
                        colorScheme="brand"
                        defaultChecked
                      />
                    </FormControl>
                  </VStack>
                </Box>
              </VStack>
            </CardBody>
          </Card>

          <Text fontSize="sm" color="gray.500" textAlign="center">
            Settings functionality is coming soon!
          </Text>
        </VStack>
      </Container>
    </Box>
  )
}

export default SettingsPage

