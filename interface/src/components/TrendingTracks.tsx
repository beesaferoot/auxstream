import {
  Box,
  Button,
  Card,
  CardBody,
  CardFooter,
  CardHeader,
  Grid,
  Image,
  useDisclosure,
} from '@chakra-ui/react'
import { Track } from '../interfaces/tracks'
import { BiLike, BiChat, BiShare } from 'react-icons/bi'
import { Assets, BASE_URL } from '../utils/constants'
import Modal from './Modal'
import AudioPlayer from './AudioPlayer'

const TrendingTracks = ({ tracks }: { tracks: Track[] }) => {
  const { isOpen, onOpen, onClose } = useDisclosure()

  return (
    <Grid
      templateColumns={[
        'repeat(1, 1fr)',
        'repeat(2, 1fr)',
        'repeat(3, 1fr)',
        'repeat(5, 1fr)',
      ]}
      gap={6}
    >
      {tracks.map((track, i) => {
        return (
          <Card
            maxW="md"
            _hover={{ shadow: 'md' }}
            _focus={{ boxShadow: 'outline' }}
            key={i}
            onClick={onOpen}
          >
            <CardHeader>{track.title}</CardHeader>
            <CardBody display="flex" flexDirection="column" alignItems="center">
              <Image
                objectFit="cover"
                alt="track cover"
                src={Assets.MusicalNote}
              ></Image>
            </CardBody>
            <CardFooter
              justify="space-between"
              flexWrap="wrap"
              sx={{
                '& > button': {
                  minW: '80px',
                },
              }}
            >
              <Button flex="1" variant="ghost" leftIcon={<BiLike />}>
                Like
              </Button>
              <Button flex="1" variant="ghost" leftIcon={<BiChat />}>
                Comment
              </Button>
              <Button flex="1" variant="ghost" leftIcon={<BiShare />}>
                Share
              </Button>
              <Modal
                title="Track"
                isOpen={isOpen}
                onClose={onClose}
                footer={`${track.title}`}
              >
                <Box
                  borderRadius="md"
                  p="4"
                  boxShadow="md"
                  overflow="hidden"
                  maxWidth="350px"
                  height={100}
                >
                  <AudioPlayer src={`${BASE_URL}serve/${track.file}`} />
                </Box>
              </Modal>
            </CardFooter>
          </Card>
        )
      })}
    </Grid>
  )
}

export default TrendingTracks
